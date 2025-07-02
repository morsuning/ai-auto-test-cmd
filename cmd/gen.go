// Package cmd 提供API自动化测试命令行工具的命令实现
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// genCmd 表示通过Dify Workflow API生成测试用例的命令
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "通过Dify Workflow API生成测试用例",
	Long: `通过Dify Workflow API生成测试用例，并保存为本地CSV文件。

示例：
  # 根据正例xml报文生成10条测试用例
  atc gen -u https://xxx.dify.com/xxx/xxx -xml -raw "xxxx" -n 10

  # 根据正例json报文生成20条测试用例
  atc gen -u https://xxx.dify.com/xxx/xxx -json -raw "xxxx" -n 20

  # （可选）根据正例xml报文和接口文档生成5条正例报文
  atc gen -u https://xxx.dify.com/xxx/xxx -xml -raw "xxxx" -n 5 -d xxx.xlsx -p`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取命令行参数
		url, _ := cmd.Flags().GetString("url")
		raw, _ := cmd.Flags().GetString("raw")
		num, _ := cmd.Flags().GetInt("num")
		doc, _ := cmd.Flags().GetString("doc")
		isXML, _ := cmd.Flags().GetBool("xml")
		isJSON, _ := cmd.Flags().GetBool("json")
		pretty, _ := cmd.Flags().GetBool("pretty")

		// 打印参数信息（仅用于框架演示）
		fmt.Println("通过Dify API生成测试用例")
		fmt.Printf("URL: %s\n", url)
		fmt.Printf("报文格式: %s\n", getFormatName(isXML, isJSON))
		fmt.Printf("原始报文: %s\n", raw)
		fmt.Printf("生成数量: %d\n", num)
		if doc != "" {
			fmt.Printf("接口文档: %s\n", doc)
		}
		fmt.Printf("美化输出: %v\n", pretty)

		// TODO: 实现通过Dify API生成测试用例的功能
		fmt.Println("功能尚未实现")
	},
}

// 获取报文格式名称
func getFormatName(isXML, isJSON bool) string {
	if isXML {
		return "XML"
	} else if isJSON {
		return "JSON"
	}
	return "未指定"
}

func init() {
	rootCmd.AddCommand(genCmd)

	// 定义命令行参数
	genCmd.Flags().StringP("url", "u", "", "目标URL（必需）")
	genCmd.Flags().StringP("raw", "r", "", "请求参数（正例报文）（必需）")
	genCmd.Flags().IntP("num", "n", 10, "生成用例数量（默认10）")
	genCmd.Flags().StringP("doc", "d", "", "接口文档路径（可选）")
	genCmd.Flags().BoolP("xml", "x", false, "使用XML格式")
	genCmd.Flags().BoolP("json", "j", false, "使用JSON格式")
	genCmd.Flags().BoolP("pretty", "p", false, "美化输出")

	// 标记必需的参数
	genCmd.MarkFlagRequired("url")
	genCmd.MarkFlagRequired("raw")
}