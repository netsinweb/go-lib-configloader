package configloader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Load は設定ファイルを読み込み、指定された型のポインタにアンマーシャリングします。
// コマンドライン引数でパスが指定されていればそれを使い、なければ実行ファイルと同じディレクトリの "Config.json" を探します。
func Load[T any](configFileName string) (*T, error) {
	// 1. 設定ファイルのパスを決定する
	if configFileName == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("could not get executable path: %w", err)
		}
		configFileName = filepath.Join(filepath.Dir(exe), "Config.json")
	}

	// 2. ファイルを読み込む
	b, err := os.ReadFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %s: %w", configFileName, err)
	}

	// 3. JSONを構造体にデコードする
	var config T
	if err := json.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("the configuration file format (JSON) is invalid: %w", err)
	}

	return &config, nil
}
