package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"super-sms-bridge/internal/config"
	"super-sms-bridge/internal/handler"
	"super-sms-bridge/internal/telegram"

	"github.com/soheilhy/cmux"
)

func startServer(addr string, handler http.Handler, isTLS bool, certFile, keyFile string) error {
	if isTLS {
		return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
	}
	return http.ListenAndServe(addr, handler)
}

func main() {
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	dataDir := flag.String("data-dir", "./data", "数据目录路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		log.Fatalf("配置验证失败: %v", err)
	}

	// 初始化Telegram客户端
	tg, err := telegram.NewClient(
		cfg.Telegram.BotToken,
		cfg.Telegram.TargetGroupID,
		*dataDir,
	)
	if err != nil {
		log.Fatalf("初始化Telegram客户端失败: %v", err)
	}

	// 初始化HTTP和WebSocket处理器
	httpHandler := handler.NewHTTPHandler(tg, cfg.HTTP.SecretKey)
	wsHandler := handler.NewWSHandler(tg, cfg.WS.SecretKey)

	// 创建路由
	mux := http.NewServeMux()
	mux.HandleFunc("/message", httpHandler.HandleMessage)
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)

	// 判断是否需要复用端口
	if cfg.HTTP.Enabled && cfg.WS.Enabled && cfg.HTTP.Port == cfg.WS.Port {
		// 使用cmux复用端口
		addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("创建监听失败: %v", err)
		}

		m := cmux.New(listener)

		// 匹配WebSocket连接
		wsListener := m.Match(cmux.HTTP1HeaderField("Upgrade", "websocket"))
		// 匹配HTTP连接
		httpListener := m.Match(cmux.Any())

		// 创建HTTP服务
		httpServer := &http.Server{
			Handler: mux,
		}

		log.Printf("在端口 %d 启动复用服务器", cfg.HTTP.Port)

		// 启动服务
		go func() {
			if err := httpServer.Serve(httpListener); err != nil {
				log.Printf("HTTP服务错误: %v", err)
			}
		}()

		go func() {
			if err := httpServer.Serve(wsListener); err != nil {
				log.Printf("WebSocket服务错误: %v", err)
			}
		}()

		// 启动cmux
		if err := m.Serve(); err != nil {
			log.Fatalf("cmux.Serve() 错误: %v", err)
		}
	} else {
		// 分别启动HTTP和WebSocket服务
		if cfg.HTTP.Enabled {
			go func() {
				addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
				log.Printf("启动HTTP服务在端口 %d", cfg.HTTP.Port)
				err := startServer(addr, mux, cfg.HTTP.TLS.Cert != "" && cfg.HTTP.TLS.Key != "",
					cfg.HTTP.TLS.Cert, cfg.HTTP.TLS.Key)
				if err != nil {
					log.Printf("HTTP服务错误: %v", err)
				}
			}()
		}

		if cfg.WS.Enabled {
			go func() {
				addr := fmt.Sprintf(":%d", cfg.WS.Port)
				log.Printf("启动WebSocket服务在端口 %d", cfg.WS.Port)
				err := startServer(addr, mux, cfg.WS.TLS.Cert != "" && cfg.WS.TLS.Key != "",
					cfg.WS.TLS.Cert, cfg.WS.TLS.Key)
				if err != nil {
					log.Printf("WebSocket服务错误: %v", err)
				}
			}()
		}

		// 保持主程序运行
		select {}
	}
}
