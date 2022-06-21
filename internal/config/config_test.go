package config

import (
	"os"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

//"RUN_ADDRESS"
//"DATABASE_URI"
//"ACCRUAL_SYSTEM_ADDRESS"

//	RunAddress           string `env:"RUN_ADDRESS"`
//	DatabaseURI          string `env:"DATABASE_URI"`
//	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`

//const (
//	runAddressDefault    = ":8081"
//	databaseURIDefault   = "postgres://gophermart:Passw0rd@localhost:5432/database_name"
//	accrualSystemAddress = "localhost:8080"
//)

func TestGetConfig(t *testing.T) {
	tests := []struct {
		name    string
		varEnvs Config
		want    Config
	}{
		{name: "Simple valid test",
			varEnvs: Config{
				"RunAddress",
				"DatabaseURI",
				"AccrualSystemAddress",
			},
			want: Config{
				"RunAddress",
				"DatabaseURI",
				"AccrualSystemAddress",
			},
		},
		{name: "Simple default test",
			varEnvs: Config{
				"",
				"",
				"",
			},
			want: Config{
				runAddressDefault,
				databaseURIDefault,
				accrualSystemAddress,
			},
		},
		{name: "Simple mix flags and Env test",
			varEnvs: Config{
				"",
				"DatabaseURI",
				"",
			},
			want: Config{
				runAddressDefault,
				"DatabaseURI",
				accrualSystemAddress,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("RUN_ADDRESS", tt.varEnvs.RunAddress)
			os.Setenv("DATABASE_URI", tt.varEnvs.DatabaseURI)
			os.Setenv("ACCRUAL_SYSTEM_ADDRESS", tt.varEnvs.AccrualSystemAddress)
			defer os.Unsetenv("RUN_ADDRESS")
			defer os.Unsetenv("DATABASE_URI")
			defer os.Unsetenv("ACCRUAL_SYSTEM_ADDRESS")

			log.Debug().Msgf("tt.varEnvs.RunAddress - %s", tt.varEnvs.RunAddress)
			log.Debug().Msgf("tt.varEnvs.DatabaseURI - %s", tt.varEnvs.DatabaseURI)
			log.Debug().Msgf("tt.varEnvs.AccrualSystemAddress - %s", tt.varEnvs.AccrualSystemAddress)

			actual := GetConfig()
			assert.Equal(t, tt.want.RunAddress, actual.RunAddress)
			assert.Equal(t, tt.want.DatabaseURI, actual.DatabaseURI)
			assert.Equal(t, tt.want.AccrualSystemAddress, actual.AccrualSystemAddress)
			//}
			//	t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			//if got := GetConfig(); !reflect.DeepEqual(got, tt.want) {
			//log.Debug().Msgf("RUN_ADDRESS - %s", os.Getenv("RUN_ADDRESS"))
			//log.Debug().Msgf("DATABASE_URI - %s", os.Getenv("DATABASE_URI"))
			//log.Debug().Msgf("ACCRUAL_SYSTEM_ADDRESS - %s", os.Getenv("ACCRUAL_SYSTEM_ADDRESS"))
		})
	}
}
