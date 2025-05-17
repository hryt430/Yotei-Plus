package logger

// Config はロガーの設定構造体
type Config struct {
	// ログレベル: "debug", "info", "warn", "error", "fatal"
	Level string `json:"level" yaml:"level"`

	// ログ出力先: "console", "file", "both"
	Output string `json:"output" yaml:"output"`

	// ファイルログ設定
	File struct {
		// ログファイルパス
		Path string `json:"path" yaml:"path"`

		// ログローテーション設定
		MaxSize    int  `json:"maxSize" yaml:"maxSize"`       // メガバイト単位
		MaxBackups int  `json:"maxBackups" yaml:"maxBackups"` // 保持する古いログファイルの最大数
		MaxAge     int  `json:"maxAge" yaml:"maxAge"`         // 日数単位
		Compress   bool `json:"compress" yaml:"compress"`     // 古いファイルを圧縮するか
	} `json:"file" yaml:"file"`

	// 開発モードか（より詳細なスタックトレースなど）
	Development bool `json:"development" yaml:"development"`
}

// DefaultConfig はデフォルトのロガー設定を返します
func DefaultConfig() *Config {
	cfg := &Config{
		Level:       "info",
		Output:      "console",
		Development: false,
	}

	cfg.File.Path = "logs/app.log"
	cfg.File.MaxSize = 100
	cfg.File.MaxBackups = 3
	cfg.File.MaxAge = 28
	cfg.File.Compress = true

	return cfg
}
