package models

import (
	"crypto/ecdsa"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
)

var lock = &sync.Mutex{}

type Config struct {
	DBConfig       DBConfig
	JWTConfig      JWTConfig
	PasswordConfig PasswordConfig
}

var config *Config

// GetConfig get singleton config and to load config from environment variables if necessary
func GetConfig() (*Config, error) {
	if config == nil {
		lock.Lock()
		defer lock.Unlock()
		if config == nil {
			dbConfig := DBConfig{}
			err := dbConfig.getConfigFromENV()
			if err != nil {
				return &Config{}, err
			}

			jwtConfig := JWTConfig{}
			err = jwtConfig.getConfigFromENV()
			if err != nil {
				return &Config{}, err
			}

			passwordConfig := PasswordConfig{}
			err = passwordConfig.getConfigFromENV()
			if err != nil {
				return &Config{}, err
			}

			config = &Config{DBConfig: dbConfig, JWTConfig: jwtConfig, PasswordConfig: passwordConfig}
		}
	}
	return config, nil
}

type DBConfig struct {
	host         string
	port         int
	user         string
	password     string
	dbname       string
	MaxOpenConns int
	MaxIdleConns int
}

// getVarFromFileOrENV load variable from a file if requested else from env
func getVarFromFileOrENV(key string) (string, error) {
	//check file name environment variable
	fileName := os.Getenv(key + "_FILE")
	if fileName != "" {
		value, err := readFile(fileName)
		if err != nil {
			return "", err
		}
		return string(value), nil
	} else {
		return os.Getenv(key), nil
	}
}

func (c *DBConfig) getConfigFromENV() error {
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return err
	}
	c.port = port

	maxOpenConns, err := strconv.Atoi(os.Getenv("MAX_OPEN_CONNS"))
	if err != nil {
		return err
	}
	c.MaxOpenConns = maxOpenConns

	maxIdleConns, err := strconv.Atoi(os.Getenv("MAX_IDLE_CONNS"))
	if err != nil {
		return err
	}
	c.MaxIdleConns = maxIdleConns

	password, err := getVarFromFileOrENV("DB_PASSWORD")
	if err != nil {
		return err
	}
	c.password = password

	user, err := getVarFromFileOrENV("DB_USER")
	if err != nil {
		return err
	}
	c.user = user

	c.host = os.Getenv("DB_HOST")
	c.dbname = os.Getenv("DB_NAME")
	return nil
}

//GetPsqlConn get postgresql formatted connection string from configuration
func (c *DBConfig) GetPsqlConn() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", c.host, c.port, c.user, c.password, c.dbname)
}

type JWTConfig struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  ecdsa.PublicKey
}

func (c *JWTConfig) getConfigFromENV() error {
	priv, err := getVarFromFileOrENV("JWT_PRIVATE_KEY")
	if err != nil {
		return err
	}

	privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(priv))
	if err != nil {
		return err
	}
	c.PrivateKey = *privateKey

	publ, err := getVarFromFileOrENV("JWT_PUBLIC_KEY")
	if err != nil {
		return err
	}

	publicKey, err := jwt.ParseECPublicKeyFromPEM([]byte(publ))
	if err != nil {
		return err
	}
	c.PublicKey = *publicKey
	return nil
}

// readFile read entire file to bytes
func readFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Panic()
		}
	}(file)
	return ioutil.ReadAll(file)
}

type PasswordConfig struct {
	BcryptCost int
	PepperKey  []byte
}

func (c *PasswordConfig) getConfigFromENV() error {
	bcryptCost, err := strconv.Atoi(os.Getenv("BCRYPT_COST"))
	if err != nil {
		return err
	}
	c.BcryptCost = bcryptCost
	pepperKey, err := getVarFromFileOrENV("PEPPER_KEY")
	if err != nil {
		return err
	}
	c.PepperKey = []byte(pepperKey)
	return err
}
