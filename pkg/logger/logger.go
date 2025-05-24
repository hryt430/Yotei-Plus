package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// instance はシングルトンロガーインスタンス
	instance *Logger
	// once はシングルトンパターン用
	once sync.Once
)

// Logger はアプリケーションのロガーラッパー
type Logger struct {
	zap  *zap.Logger
	atom zap.AtomicLevel
	cfg  *Config
}

// Init はロガーを初期化します
func Init(cfg *Config) {
	once.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}
		instance = NewLogger(cfg)
	})
}

// Get はシングルトンロガーインスタンスを返します
func Get() *Logger {
	if instance == nil {
		Init(nil)
	}
	return instance
}

// newLogger は新しいロガーインスタンスを作成します
func NewLogger(cfg *Config) *Logger {
	// デフォルトレベルはInfo
	level := zap.InfoLevel

	// 設定からログレベルを解析
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	}

	// 動的にレベルを変更できるようにAtomicLevelを使用
	atom := zap.NewAtomicLevelAt(level)

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 開発モード用の設定
	if cfg.Development {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// コア設定の作成
	var cores []zapcore.Core

	// コンソール出力の設定
	if cfg.Output == "console" || cfg.Output == "both" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), atom)
		cores = append(cores, consoleCore)
	}

	// ファイル出力の設定
	if cfg.Output == "file" || cfg.Output == "both" {
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)
		fileWriteSyncer := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.File.Path,
			MaxSize:    cfg.File.MaxSize,    // メガバイト
			MaxBackups: cfg.File.MaxBackups, // ファイル数
			MaxAge:     cfg.File.MaxAge,     // 日数
			Compress:   cfg.File.Compress,   // 圧縮するか
		})
		fileCore := zapcore.NewCore(fileEncoder, fileWriteSyncer, atom)
		cores = append(cores, fileCore)
	}

	// すべてのコアを結合
	core := zapcore.NewTee(cores...)

	// ロガーオプションの設定
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	}

	if cfg.Development {
		options = append(options, zap.Development())
	}

	// zapロガーの作成
	zapLogger := zap.New(core, options...)

	return &Logger{
		zap:  zapLogger,
		atom: atom,
		cfg:  cfg,
	}
}

// SetLevel はログレベルを動的に変更します
func (l *Logger) SetLevel(level string) {
	switch level {
	case "debug":
		l.atom.SetLevel(zap.DebugLevel)
	case "info":
		l.atom.SetLevel(zap.InfoLevel)
	case "warn":
		l.atom.SetLevel(zap.WarnLevel)
	case "error":
		l.atom.SetLevel(zap.ErrorLevel)
	case "fatal":
		l.atom.SetLevel(zap.FatalLevel)
	default:
		l.Warn("Unknown log level: " + level + ", using info instead")
		l.atom.SetLevel(zap.InfoLevel)
	}
}

// GetLevel は現在のログレベルを文字列で返します
func (l *Logger) GetLevel() string {
	level := l.atom.Level()
	switch level {
	case zap.DebugLevel:
		return "debug"
	case zap.InfoLevel:
		return "info"
	case zap.WarnLevel:
		return "warn"
	case zap.ErrorLevel:
		return "error"
	case zap.FatalLevel:
		return "fatal"
	default:
		return fmt.Sprintf("unknown(%v)", level)
	}
}

// Debug はデバッグメッセージをログに記録します
func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.zap.Debug(msg, fields...)
}

// Info は情報メッセージをログに記録します
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.zap.Info(msg, fields...)
}

// Warn は警告メッセージをログに記録します
func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	l.zap.Warn(msg, fields...)
}

// Error はエラーメッセージをログに記録します
func (l *Logger) Error(msg string, fields ...zapcore.Field) {
	l.zap.Error(msg, fields...)
}

// Fatal は致命的なエラーメッセージをログに記録し、プログラムを終了します
func (l *Logger) Fatal(msg string, fields ...zapcore.Field) {
	l.zap.Fatal(msg, fields...)
}

// With は追加のフィールドを持つ新しいロガーを返します
func (l *Logger) With(fields ...zapcore.Field) *Logger {
	return &Logger{
		zap:  l.zap.With(fields...),
		atom: l.atom,
		cfg:  l.cfg,
	}
}

// Field ヘルパー関数群
func String(key, value string) zapcore.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) zapcore.Field {
	return zap.Int64(key, value)
}

func Float64(key string, value float64) zapcore.Field {
	return zap.Float64(key, value)
}

func Bool(key string, value bool) zapcore.Field {
	return zap.Bool(key, value)
}

func Any(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}

func Error(err error) zapcore.Field {
	return zap.Error(err)
}

// Sync はバッファされたログをすべてフラッシュします
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// Close はロガーリソースを解放します
func (l *Logger) Close() error {
	return l.Sync()
}
