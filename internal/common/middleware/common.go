package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger はリクエストのロギングを行うミドルウェアです
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// リクエストのハンドリング
		next.ServeHTTP(w, r)

		// リクエスト処理後のロギング
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			duration,
		)
	})
}

// Recovery はパニックからの回復を処理するミドルウェアです
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("パニックが発生しました: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// CORS はCross-Origin Resource Sharingを処理するミドルウェアです
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORSヘッダーの設定
		w.Header().Set("Access-Control-Allow-Origin", "*") // 本番環境では適切なオリジンを設定
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// プリフライトリクエスト（OPTIONS）の処理
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
