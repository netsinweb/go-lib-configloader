# go-lib-configloader(configloader)

## Description

- go-lib-configloaderモジュール(configloaderパッケージ)は設定用JSONファイルの読み込み、と構造体への展開を行います。
- 使用前に設定用構造体の定義を行ってください。
- 設定用構造体の定義にしたがってJSONの展開を行います。
- Load()メソッドを使用して読み込みと展開を行いますが、引数に設定ファイルのパスを指定できます。
- パスが未指定(引数が"")の場合は、実行ファイルと同じディレクトリ上にある「Config.json」を読み込みます。

## Usage

~~~sh
go get https://github.com/netsinweb/go-lib-configloader@v1.0.0
~~~

## Example

### 組み込み例

main.go

~~~go
package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/netsinweb/go-lib-configloader"
)

type config struct {
	Web WebConfig       `json:"Web"`
}

type WebConfig struct {
	Port string `json:"Port"`
}

func main() {
	// Flag
	conf := flag.String("conf", "", "Config file path (ex: \"Config.json\")")
	flag.Parse()

	// Config
	appConfig, err := configloader.Load[config](*conf)
	if err != nil {
		slog.Error("Failed to load configuration", slog.String("Error", err.Error()))
		os.Exit(1)
	}

	// appConfig.WebConfig 等で設定を取り出して使用
}
~~~

Config.json

~~~json
{
  "Web": {
    "Port": "8080"
  }
}
~~~

StartUp

~~~sh
go run . -conf="Config.json"

app.exe -conf="Config.json"

# 引数無しの場合はapp.exeと同じ階層にあるConfig.jsonを読み込む
app.exe
~~~

### Web application background(Template)

mian.go

~~~go
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/netsinweb/go-lib-configloader"
	postgres "github.com/netsinweb/go-lib-db-postgres"
	"github.com/netsinweb/go-lib-logger"
)

type config struct {
	Web WebConfig       `json:"Web"`
	Db  postgres.Config `json:"Db"`
	Log logger.Config   `json:"Log"`
}

type WebConfig struct {
	Port string `json:"Port"`
}

func main() {
	// Flag
	conf := flag.String("conf", "", "Config file path (ex: \"Config.json\")")
	flag.Parse()

	// Config
	appConfig, err := configloader.Load[config](*conf)
	if err != nil {
		slog.Error("Failed to load configuration", slog.String("Error", err.Error()))
		os.Exit(1)
	}

	// Logger
	logCloser, err := logger.Setup(appConfig.Log)
	if err != nil {
		slog.Error("Failed to setup logger", slog.String("Error", err.Error()))
		os.Exit(1)
	}
	defer logCloser.Close()

	// Database
	dbi, err := postgres.NewDB(appConfig.Db)
	if err != nil {
		slog.Error("Could not connect to ois database", slog.String("Status", "error"), slog.String("Error", err.Error()))
		os.Exit(1)
	}
	defer dbi.Close()

	dbConnect := dbi.GetDB()
	slog.Info("Connected to ois database", slog.String("Status", "success"))

	// DI
	// :

	// Routing
	router := gin.Default()
	{
		api := router.Group("/api")
		api.GET("/ping", func(gc *gin.Context) {
			gc.JSON(http.StatusOK, gin.H{"message":"ping"})
		})
	}
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", appConfig.Web.Port),
		Handler: router,
	}

	// Startup Server
	go func() {
		slog.Info("Server is starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to start server", slog.String("Error", err.Error()))
			os.Exit(1)
		}
	}()

	// Shutdown Server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Server is shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", slog.String("Error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Server exiting")
}
~~~

Config.json

~~~json
{
  "Web": {
      "port": "8080"
  },
  "Db": {
      "Host": "127.0.0.1",
      "User": "user",
      "Password": "password",
      "Database": "applicationdata",
      "Port": "5432"
  },
  "Log": {
      "Path": "./log",
      "File": "application.log"
  }
}
~~~
