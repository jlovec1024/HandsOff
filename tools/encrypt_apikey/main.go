package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/handsoff/handsoff/pkg/crypto"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	godotenv.Load()

	// 命令行参数
	apiKey := flag.String("key", "", "API Key to encrypt (required)")
	encryptionKey := flag.String("encryption-key", os.Getenv("ENCRYPTION_KEY"), "Encryption key (default: from .env)")
	decrypt := flag.String("decrypt", "", "Encrypted text to decrypt")
	flag.Parse()

	// 如果是解密模式
	if *decrypt != "" {
		decryptMode(*decrypt, *encryptionKey)
		return
	}

	// 加密模式
	if *apiKey == "" {
		fmt.Println("❌ Error: API Key is required")
		fmt.Println("\nUsage:")
		fmt.Println("  加密: go run tools/encrypt_apikey/main.go -key 'sk-xxxxxxxx'")
		fmt.Println("  解密: go run tools/encrypt_apikey/main.go -decrypt 'encrypted-text'")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *encryptionKey == "" {
		log.Fatal("❌ Error: Encryption key not found (set ENCRYPTION_KEY in .env or use -encryption-key flag)")
	}

	encryptMode(*apiKey, *encryptionKey)
}

// encryptMode 加密模式
func encryptMode(apiKey, encryptionKey string) {
	fmt.Println("==============================================")
	fmt.Println("API Key 加密工具")
	fmt.Println("==============================================\n")

	// 加密
	encrypted, err := crypto.EncryptString(apiKey, encryptionKey)
	if err != nil {
		log.Fatalf("❌ 加密失败: %v", err)
	}

	fmt.Println("✅ 加密成功\n")
	fmt.Println("原始 API Key:")
	fmt.Printf("  %s\n\n", apiKey)

	fmt.Println("加密后的值 (用于数据库存储):")
	fmt.Printf("  %s\n\n", encrypted)

	fmt.Println("==============================================")
	fmt.Println("使用方法:")
	fmt.Println("==============================================")
	fmt.Println("1. 复制上面的加密值")
	fmt.Println("2. 在数据库中更新 llm_providers.api_key 或 git_platform_configs.access_token")
	fmt.Println("\nSQL 示例:")
	fmt.Printf("  UPDATE llm_providers SET api_key = '%s' WHERE id = 1;\n\n", encrypted)

	// 验证解密
	fmt.Println("验证解密:")
	decrypted, err := crypto.DecryptString(encrypted, encryptionKey)
	if err != nil {
		log.Fatalf("❌ 解密验证失败: %v", err)
	}

	if decrypted == apiKey {
		fmt.Println("  ✅ 解密验证成功")
	} else {
		fmt.Println("  ❌ 解密验证失败 (值不匹配)")
	}
}

// decryptMode 解密模式
func decryptMode(encrypted, encryptionKey string) {
	fmt.Println("==============================================")
	fmt.Println("API Key 解密工具")
	fmt.Println("==============================================\n")

	if encryptionKey == "" {
		log.Fatal("❌ Error: Encryption key not found")
	}

	// 解密
	decrypted, err := crypto.DecryptString(encrypted, encryptionKey)
	if err != nil {
		log.Fatalf("❌ 解密失败: %v", err)
	}

	fmt.Println("✅ 解密成功\n")
	fmt.Println("加密值:")
	fmt.Printf("  %s\n\n", encrypted)

	fmt.Println("解密后的 API Key:")
	fmt.Printf("  %s\n\n", decrypted)
}
