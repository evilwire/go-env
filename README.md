Golang Marshaler from Environment Variables
===========================================

[![Build Status](https://travis-ci.org/evilwire/go-env.svg?branch=master)](https://travis-ci.org/evilwire/go-env)
[![GoDoc](https://godoc.org/github.com/evilwire/go-env?status.svg)](https://godoc.org/github.com/evilwire/go-env)

Golang (1.5+) package for marshalling objects from environment variable values.
There are many like packages. The one that inspires this API the most is
the `json` package. The idea is that configuration objects are stored
as environment variables, especially when running as a containerised
application.

### Example

The following is a full and well contrived application that showcases
how to use the API to retrieve data from environments. The application

- retrieves a list of things to say and wait time (via config)
- for everything to be said, it prints a simple statement

Disclaimer: this is a silly toy application. Don't use this as an example
of good application design.

```go
package main


import (
    "fmt"
    "github.com/evilwire/go-env"
    "time"
)


type SomeConfig struct {
        User string `env:"USER_NAME"`
        
        Application struct {
                WaitDuration time.Duration `env:"WAIT_DURATION"`
                
                // this is best obtained from flags, but you can equally
                // retrieve this from environment variables
                ThingsToSay []string `env:"THINGS_TO_SAY"`
        } `env:"APP_"`
}


// Sets up my application by reading the configs
func setup(env goenv.EnvReader) (*SomeConfig, error) {
        // create a Marshaler using our lovely default, which knows
        // how to marshal a set of things
        marshaler := goenv.DefaultEnvMarshaler{
                env,
        }
        
        // instantiate an empty config 
        config := SomeConfig{} 
        err := marshaler.Unmarshal(&config)
        if err != nil {
            return nil, err
        }
        
        return &config, nil
}


func main() {
        config, err := setup(goenv.NewOsEnvReader())
        if err != nil {
                panic(err)
        }
        
        for _, line := range config.Application.ThingsToSay {
                fmt.Printf("%s says: ", config.User)
                fmt.Println(line)
                time.Sleep(config.Application.WaitDuration)
        }
        
        fmt.Println("We're done!")
}

```

Compile your application say `silly-app`, and run your application

```sh
USER_NAME='Michael Bluth' \
APP_WAIT_DURATION=2s \
APP_THINGS_TO_SAY='hiya,how are you,bye' /path/to/silly-app
```

### Expanding the API

One of the cases that I encounter is using [AWS KMS](https://aws.amazon.com/kms/) to manage
secrets. For example, you may want to store KMS-encrypted credentials in a distributed 
version control source such GitHub, and passed into your application directly via 
environment variables.

```go
```
