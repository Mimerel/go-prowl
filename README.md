# go-prowl
Send notification to Prowl

This go software is designed to send prowl Request to you prowl token.
It can also store the prown events in your elasticsearch if you have one set.

* Prerequisits

Having a prowl token -> visit the prowl website to create your account.
Have GO setup on you system


* Using the app.

The app will be running on the port set in the configuration.yaml file.
you need to set an environment variable to specify the full path to you configuration file.

the environment variable is : LOGGER_CONFIGURATION_FILE

If no env variable is set, the application will search for it in the current file from 
which is started the application or the following path : 
```
/home/pi/go/src/go-prowl/configuration.yaml
```

* RUN : to run the application either : 
```
go run main.go
```
or
```
go build  // to build the application
```
then 
```
./go-prowl // to run the build
```

You will probably be missing dependencies

to add a dependency run 

```
go get <name of dependency>
```

for example : 
```
go get github.com/Mimerel/go-logger-client
```

* Configuration file

```
token: <your prowl token>
elasticSearch:
    url: <ip to your elasticsearch>:9200
host: <name of the collection in your elasticsearch that will be used to store the events>
port: 9999  // port on which will run this application

// this part enables you to ignore (no send, no store) any events during certain periods
// from : 2155 means as from 21:55 / 9:55pm 
ignore:
  - from: 2155
    to: 2156
  - from: 2159
    to: 2200
```