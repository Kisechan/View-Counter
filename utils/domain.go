package utils

import (
	"net/http"
	"strings"
)

func ExtractDomain(r *http.Request) string {
	// 优先从 Referer 获取域名
	referer := r.Header.Get("Referer")
	if referer != "" {
		if strings.HasPrefix(referer, "http://") || strings.HasPrefix(referer, "https://") {
			host := strings.TrimPrefix(strings.TrimPrefix(referer, "http://"), "https://")
			host = strings.Split(host, "/")[0]
			host = strings.Split(host, ":")[0]
			return strings.ToLower(host)
		}
	}
	
	// 使用 X-Real-Host 和 Host头部检测域名
	host := r.Header.Get("X-Real-Host")
	if host == "" {
		host = r.Host
	}
	return strings.ToLower(strings.Split(host, ":")[0])
}