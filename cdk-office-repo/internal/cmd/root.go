/*
 * MIT License
 *
 * Copyright (c) 2025 CDK-Office
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package cmd

import (
	"log"

	"github.com/linux-do/cdk-office/internal/config"
	"github.com/linux-do/cdk-office/internal/db/migrator"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "cdk-office",
	Version: "1.0.0",
	Short:   "CDK-Office企业内容管理平台",
	Long: `CDK-Office是一个集成了Dify AI平台的企业内容管理平台，
支持智能文档管理、AI问答和知识库管理功能。`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// 加载配置
		config.Load()

		// 执行数据库迁移
		migrator.Migrate()
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalf("[CMD] please provide a command\n")
		}
		appMode := args[0]
		switch appMode {
		case "api":
			apiCmd.Run(apiCmd, args)
		case "scheduler":
			schedulerCmd.Run(schedulerCmd, args)
		case "worker":
			workerCmd.Run(workerCmd, args)
		default:
			log.Fatal("[CMD] unknown app mode\n")
		}
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// Execute 执行命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("[CMD] execute failed; %s\n", err)
	}
}
