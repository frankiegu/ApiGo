package apigoconfig

import (
	"bytes"
	"io/ioutil"
	"log"

	"github.com/kardianos/osext"
	"github.com/spf13/viper"
)

//Get ...
func Get() (*viper.Viper, error) {
	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatal(err)
	}
	v := viper.New()

	v.SetDefault("logFolder", "${TEMP}")
	v.SetDefault("JobsDatabase.AdminDatabase", "postgres")
	v.SetDefault("JobsDatabase.Driver", "postgres")
	v.SetDefault("JobsDatabase.ConnectionString", "Host=localhost;Port=5432;Database=agdatabase;Username=aguser;Password=agpassword;SSL Mode=Require;Trust Server Certificate=true")
	v.SetDefault("Bindings", "0.0.0.0:1203")
	v.SetDefault("RoutesConfigPath", "config/routes")
	v.SetDefault("Debug", true)

	v.SetConfigName("default")
	v.SetConfigType("json")
	v.AddConfigPath("config")
	v.AddConfigPath(".")
	v.AddConfigPath(folderPath)
	v.AddConfigPath(folderPath + "/config")
	err = v.ReadInConfig()
	if err != nil {
		log.Printf("couldn't read the configuration from file: \n%v\n", err)
	}
	return v, nil
}

//GetRouteConfig ...
func GetRouteConfig(pathFile string) *viper.Viper {
	v := viper.New()
	v.SetConfigType("json")
	log.Printf("Read file '%s'", pathFile)
	f, err := ioutil.ReadFile(pathFile)
	log.Printf("Load configuration from content of file '%s'", pathFile)
	err = v.ReadConfig(bytes.NewBuffer(f))
	if err != nil {
		log.Printf("couldn't read the configuration from file: \n%v\n", err)
	}
	return v
}
