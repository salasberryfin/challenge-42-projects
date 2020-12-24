package main

import (
    // "flag"
    "log"
    "fmt"
    "net/http"
    "gopkg.in/yaml.v2"
    "os"
    "path"
     "time"
)

// const config_file_yml = "conf.yml"
const config_file_yml = "server-defaults/conf.yml"

type Method struct {
    Endpoint   string `yaml:"endpoint"`
    Type        string `yaml:"type"`
    SourceFile  string `yaml:"source_file"`
    Error       string `yaml:"error"`
}

type ServerConfiguration struct {
    ServerPort      string `yaml:"serverport"`
    Methods         []Method `yaml:"methods"`
}

type logWriter struct {
}

// Custom logger format
func (writer logWriter) Write(bytes []byte) (int, error) {
    return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + string(bytes))
}

func readYmlConfFile(filename *os.File) ServerConfiguration {
    decoder := yaml.NewDecoder(filename)
    configuration := ServerConfiguration{}
    err := decoder.Decode(&configuration)
    if err != nil {
        log.Println("Error", err)
    }

    return configuration
}

func methodGenerator(method Method) error {
    definition := func(w http.ResponseWriter, r *http.Request) {
        log.SetOutput(new(logWriter))
        log.Printf("[%q] - You tried to reach %q.", r.Method, method.Endpoint)
        switch r.Method {
        case http.MethodGet:
            log.Println("Detected a GET request.")
            http.ServeFile(w, r, method.SourceFile)
        default:
            log.Printf("Method %q is not supported.", r.Method)
            http.ServeFile(w, r, method.Error)
        }
    }
    http.HandleFunc(fmt.Sprintf("/%s", method.Endpoint), definition)

    return nil
}

func createDynamicMethods(config ServerConfiguration) error {
    for index, method := range config.Methods {
        log.Printf("Method %d: %q\n", index, method.Endpoint)
        // Generate each endpoint dynamically
        err := methodGenerator(method)
        if err != nil {
            log.Fatal(err)
        }
    }

    return nil
}

func main() {
    wd, err := os.Getwd()
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Current work directory is: ", wd)
    yml_file, _ := os.Open(path.Join(wd, config_file_yml))
    defer yml_file.Close()

    log.Println("YML file is: ", config_file_yml)

    yml_conf := readYmlConfFile(yml_file)

    log.Println("Starting server on port ", yml_conf.ServerPort)

    log.Printf("Creating methods from %q...", config_file_yml)

    err = createDynamicMethods(yml_conf)
    if err != nil {
        log.Fatal(err)
    }

    http.ListenAndServe(":" + yml_conf.ServerPort, nil)
}

