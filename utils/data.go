/*
Copyright © 2025 API自动化测试命令行工具

*/
package utils

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// 保存原始字段顺序和类型信息
var originalKeyOrder []string
var originalValueTypes map[string]string

func init() {
	// 初始化随机数生成器
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// ParseXML 解析XML字符串为map[string]interface{}
func ParseXML(xmlStr string) (map[string]interface{}, error) {
	// 使用自定义的XML解析函数
	result, err := XMLToMap(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("解析XML失败: %v", err)
	}
	
	// 提取XML字段顺序
	// 由于XML解析过程中字段顺序可能已经丢失，我们尝试从原始XML字符串中提取
	keys := extractXMLKeys(xmlStr)
	if len(keys) > 0 {
		originalKeyOrder = keys
	} else {
		// 如果无法从原始字符串提取，则使用解析后的结果的键
		keys = make([]string, 0, len(result))
		for key := range result {
			keys = append(keys, key)
		}
		originalKeyOrder = keys
	}
	
	// 如果结果中有根元素，提取其内容作为实际数据
	for _, rootValue := range result {
		if rootMap, ok := rootValue.(map[string]interface{}); ok {
			// 使用根元素内的数据替换整个结果
			return rootMap, nil
		}
	}
	
	return result, nil
}

// extractXMLKeys 从XML字符串中提取字段顺序
func extractXMLKeys(xmlStr string) []string {
	// 简单的XML解析，提取顶层元素
	var keys []string
	var seenKeys = make(map[string]bool)
	
	// 使用正则表达式提取标签
	tagRegex := regexp.MustCompile(`<([^\s>/]+)[^>]*>([^<]*)</([^\s>/]+)>|<([^\s>/]+)[^>]*/>|<([^\s>/]+)[^>]*>`) 
	matches := tagRegex.FindAllStringSubmatch(xmlStr, -1)
	
	for _, match := range matches {
		var tagName string
		if match[1] != "" {
			tagName = match[1] // 完整标签 <tag>content</tag>
		} else if match[4] != "" {
			tagName = match[4] // 自闭合标签 <tag/>
		} else if match[5] != "" {
			tagName = match[5] // 开始标签 <tag>
		}
		
		// 忽略XML声明和根元素
		if tagName != "" && tagName != "?xml" && !strings.Contains(tagName, ":") {
			// 避免重复添加
			if !seenKeys[tagName] {
				keys = append(keys, tagName)
				seenKeys[tagName] = true
			}
		}
	}
	
	return keys
}

// XMLToMap 将XML字符串转换为map
func XMLToMap(xmlStr string) (map[string]interface{}, error) {
	// 创建一个自定义的解码器
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))
	decoder.Strict = false

	// 使用一个临时结构体来存储XML数据
	type XMLNode struct {
		XMLName xml.Name
		Content []byte     `xml:",chardata"`
		Attrs   []xml.Attr `xml:",any,attr"`
		Nodes   []XMLNode  `xml:",any"`
	}

	var node XMLNode
	if err := decoder.Decode(&node); err != nil {
		return nil, err
	}

	// 将XMLNode转换为map
	result := make(map[string]interface{})
	result[node.XMLName.Local] = nodeToMap(node)

	// 简化结果，提取实际内容
	return simplifyXMLMap(result), nil
}

// simplifyXMLMap 简化XML转换后的map结构，提取实际内容
func simplifyXMLMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 处理根节点
	for key, value := range data {
		if key == "_name" || key == "@attrs" {
			continue // 跳过节点名称和属性
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// 如果是简单的内容节点 (只有#content和_name)
			if content, ok := v["#content"]; ok {
				// 尝试将内容转换为数值类型
				strContent, isStr := content.(string)
				if isStr {
					// 尝试转换为整数
					if intVal, err := strconv.ParseInt(strContent, 10, 64); err == nil {
						result[key] = intVal
						continue
					}
					// 尝试转换为浮点数
					if floatVal, err := strconv.ParseFloat(strContent, 64); err == nil {
						result[key] = floatVal
						continue
					}
					// 尝试转换为布尔值
					if boolVal, err := strconv.ParseBool(strContent); err == nil {
						result[key] = boolVal
						continue
					}
				}
				result[key] = content
			} else if len(v) == 1 && v["_name"] != nil {
				// 空节点，跳过
				continue
			} else {
				// 递归处理子节点
				simplified := simplifyXMLMap(v)
				if len(simplified) > 0 {
					result[key] = simplified
				}
			}
		case []interface{}:
			// 处理数组
			array := make([]interface{}, 0, len(v))
			for _, item := range v {
				if mapItem, ok := item.(map[string]interface{}); ok {
					// 如果数组元素是简单的内容节点
					if content, ok := mapItem["#content"]; ok {
						// 尝试将内容转换为数值类型
						strContent, isStr := content.(string)
						if isStr {
							// 尝试转换为整数
							if intVal, err := strconv.ParseInt(strContent, 10, 64); err == nil {
								array = append(array, intVal)
								continue
							}
							// 尝试转换为浮点数
							if floatVal, err := strconv.ParseFloat(strContent, 64); err == nil {
								array = append(array, floatVal)
								continue
							}
							// 尝试转换为布尔值
							if boolVal, err := strconv.ParseBool(strContent); err == nil {
								array = append(array, boolVal)
								continue
							}
						}
						array = append(array, content)
					} else {
						// 递归处理复杂节点
						simplified := simplifyXMLMap(mapItem)
						if len(simplified) > 0 {
							array = append(array, simplified)
						}
					}
				} else if item != nil {
					array = append(array, item)
				}
			}
			if len(array) > 0 {
				result[key] = array
			}
		default:
			if v != nil {
				result[key] = v
			}
		}
	}

	return result
}

// nodeToMap 将XMLNode转换为map
func nodeToMap(node interface{}) interface{} {
	// 使用反射处理XMLNode结构
	v := reflect.ValueOf(node)
	
	// 处理不同类型的节点
	switch v.Kind() {
	case reflect.String:
		return v.String()
		
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
		
	case reflect.Float32, reflect.Float64:
		return v.Float()
		
	case reflect.Bool:
		return v.Bool()
		
	case reflect.Slice:
		// 处理字节数组
		if v.Type().Elem().Kind() == reflect.Uint8 {
			// 将[]byte转换为字符串并尝试解析
			s := strings.TrimSpace(string(v.Bytes()))
			if s == "" {
				return nil
			}
			
			// 尝试解析为整数
			if i, err := strconv.ParseInt(s, 10, 64); err == nil {
				return i
			}
			
			// 尝试解析为浮点数
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
			
			// 尝试解析为布尔值
			if b, err := strconv.ParseBool(s); err == nil {
				return b
			}
			
			// 否则返回字符串
			return s
		}
		
		// 处理其他类型的切片
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = nodeToMap(v.Index(i).Interface())
		}
		return result
		
	case reflect.Map:
		// 处理map
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			strKey := fmt.Sprintf("%v", key.Interface())
			result[strKey] = nodeToMap(v.MapIndex(key).Interface())
		}
		return result
		
	case reflect.Struct:
		// 处理XMLNode结构
		result := make(map[string]interface{})
		
		// 获取XMLName字段
		xmlNameField := v.FieldByName("XMLName")
		if xmlNameField.IsValid() {
			xmlName := xmlNameField.Interface().(xml.Name)
			result["_name"] = xmlName.Local
		}
		
		// 获取Content字段
		contentField := v.FieldByName("Content")
		if contentField.IsValid() && contentField.Len() > 0 {
			content := strings.TrimSpace(string(contentField.Bytes()))
			if content != "" {
				result["#content"] = content
			}
		}
		
		// 获取Attrs字段
		attrsField := v.FieldByName("Attrs")
		if attrsField.IsValid() && attrsField.Len() > 0 {
			attrs := make(map[string]interface{})
			for i := 0; i < attrsField.Len(); i++ {
				attr := attrsField.Index(i).Interface().(xml.Attr)
				attrs[attr.Name.Local] = attr.Value
			}
			result["@attrs"] = attrs
		}
		
		// 获取Nodes字段
		nodesField := v.FieldByName("Nodes")
		if nodesField.IsValid() && nodesField.Len() > 0 {
			childNodes := make(map[string]interface{})
			for i := 0; i < nodesField.Len(); i++ {
				childNode := nodesField.Index(i).Interface()
				childNodeValue := reflect.ValueOf(childNode)
				xmlNameField := childNodeValue.FieldByName("XMLName")
				if !xmlNameField.IsValid() {
					continue
				}
				
				xmlName := xmlNameField.Interface().(xml.Name)
				name := xmlName.Local
				value := nodeToMap(childNode)
				
				// 检查是否已存在同名节点
				if existing, ok := childNodes[name]; ok {
					// 如果已存在，转换为数组
					switch v := existing.(type) {
					case []interface{}:
						childNodes[name] = append(v, value)
					default:
						childNodes[name] = []interface{}{v, value}
					}
				} else {
					childNodes[name] = value
				}
			}
			
			// 合并子节点到结果
			for k, v := range childNodes {
				result[k] = v
			}
		}
		
		return result
		
	default:
		// 处理其他类型
		return fmt.Sprintf("%v", v.Interface())
	}
}

// convertNumbers 将json.Number转换为适当的数字类型
func convertNumbers(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		switch v := value.(type) {
		case json.Number:
			// 尝试转换为整数
			if intVal, err := v.Int64(); err == nil {
				// 检查是否在int范围内
				if intVal >= math.MinInt32 && intVal <= math.MaxInt32 {
					result[key] = int(intVal)
				} else {
					result[key] = intVal
				}
			} else {
				// 转换为浮点数
				if floatVal, err := v.Float64(); err == nil {
					result[key] = floatVal
				} else {
					// 保持原始字符串
					result[key] = string(v)
				}
			}
		case map[string]interface{}:
			// 递归处理嵌套对象
			result[key] = convertNumbers(v)
		case []interface{}:
			// 处理数组
			arr := make([]interface{}, len(v))
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					arr[i] = convertNumbers(itemMap)
				} else if itemNum, ok := item.(json.Number); ok {
					if intVal, err := itemNum.Int64(); err == nil {
						if intVal >= math.MinInt32 && intVal <= math.MaxInt32 {
							arr[i] = int(intVal)
						} else {
							arr[i] = intVal
						}
					} else if floatVal, err := itemNum.Float64(); err == nil {
						arr[i] = floatVal
					} else {
						arr[i] = string(itemNum)
					}
				} else {
					arr[i] = item
				}
			}
			result[key] = arr
		default:
			result[key] = value
		}
	}
	return result
}

// ParseJSON 解析JSON字符串并保留字段顺序
func ParseJSON(jsonStr string) (map[string]interface{}, error) {
	// 创建一个空接口来存储解析结果
	var result map[string]interface{}

	// 使用Decoder来保持数字的原始格式
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	// 转换json.Number为适当的类型
	result = convertNumbers(result)

	// 记录字段顺序
	// 由于Go的标准json包不保留字段顺序，我们需要手动提取顺序
	// 通过简单的字符串处理来提取字段顺序
	keys := extractJSONKeys(jsonStr)
	if len(keys) > 0 {
		// 确保字段顺序正确
		originalKeyOrder = keys
		fmt.Printf("JSON字段顺序: %v\n", originalKeyOrder)
	}

	return result, nil
}

// extractJSONKeys 从JSON字符串中提取字段顺序
func extractJSONKeys(jsonStr string) []string {
	// 移除所有空白字符，简化处理
	jsonStr = strings.ReplaceAll(jsonStr, " ", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\n", "")
	jsonStr = strings.ReplaceAll(jsonStr, "\t", "")
	
	// 确保是一个对象
	if !strings.HasPrefix(jsonStr, "{") || !strings.HasSuffix(jsonStr, "}") {
		return nil
	}
	
	// 移除首尾的花括号
	jsonStr = jsonStr[1 : len(jsonStr)-1]
	
	// 提取字段名
	var keys []string
	var inQuote bool
	var inArray int
	var inObject int
	var start int
	var key string
	
	for i := 0; i < len(jsonStr); i++ {
		ch := jsonStr[i]
		
		// 处理引号
		if ch == '"' && (i == 0 || jsonStr[i-1] != '\\') {
			inQuote = !inQuote
			if inQuote && inArray == 0 && inObject == 0 {
				// 可能是字段名的开始
				start = i + 1
			} else if !inQuote && inArray == 0 && inObject == 0 {
				// 可能是字段名的结束
				key = jsonStr[start:i]
			}
		}
		
		// 处理冒号（字段名和值的分隔符）
		if ch == ':' && !inQuote && inArray == 0 && inObject == 0 && key != "" {
			keys = append(keys, key)
			key = ""
		}
		
		// 处理数组
		if ch == '[' && !inQuote {
			inArray++
		} else if ch == ']' && !inQuote {
			inArray--
		}
		
		// 处理对象
		if ch == '{' && !inQuote {
			inObject++
		} else if ch == '}' && !inQuote {
			inObject--
		}
		
		// 处理逗号（值的分隔符）
		if ch == ',' && !inQuote && inArray == 0 && inObject == 0 {
			// 重置，准备处理下一个字段
			key = ""
		}
	}
	
	return keys
}

// GenerateTestCases 根据原始数据生成测试用例
func GenerateTestCases(data map[string]interface{}, count int) []map[string]interface{} {
	testCases := make([]map[string]interface{}, count)
	
	// 记录原始字段顺序
	keys := make([]string, 0, len(data))
	// 按照输入顺序记录字段
	for key := range data {
		keys = append(keys, key)
	}
	
	// 记录原始数据类型
	types := make(map[string]string)
	for key, value := range data {
		switch v := value.(type) {
		case int, int8, int16, int32, int64:
			types[key] = "int"
		case float32, float64:
			types[key] = "float"
			// 记录精度
			strVal := fmt.Sprintf("%v", v)
			if dotIndex := strings.Index(strVal, "."); dotIndex != -1 {
				precision := len(strVal) - dotIndex - 1
				types[key] = fmt.Sprintf("float:%d", precision)
			}
		case string:
			// 检查是否是逗号分隔的数组（XML中的数组表示）
			if strings.Contains(v, ",") {
				// 可能是XML中的数组
				types[key] = "array-string"
			} else if _, err := strconv.ParseInt(v, 10, 64); err == nil {
				types[key] = "int-string"
			} else if _, err := strconv.ParseFloat(v, 64); err == nil {
				// 记录精度
				if dotIndex := strings.Index(v, "."); dotIndex != -1 {
					precision := len(v) - dotIndex - 1
					types[key] = fmt.Sprintf("float-string:%d", precision)
				} else {
					types[key] = "float-string:0"
				}
			} else {
				types[key] = "string"
			}
		case bool:
			types[key] = "bool"
		case []interface{}:
			types[key] = "array"
		case map[string]interface{}:
			types[key] = "object"
		default:
			types[key] = "unknown"
		}
	}
	
	// 生成指定数量的测试用例
	for i := 0; i < count; i++ {
		testCase := make(map[string]any)
		// 按原始顺序处理每个字段
		for _, key := range keys {
			testCase[key] = generateVariation(data[key], 0.5) // 上下浮动50%
		}
		testCases[i] = testCase
	}
	
	// 保存原始字段顺序和类型信息到全局变量
	originalKeyOrder = keys
	originalValueTypes = types
	
	return testCases
}

// generateVariation 根据原始值生成变化值
func generateVariation(value interface{}, variationRate float64) interface{} {
	// 根据值的类型进行不同处理
	switch v := value.(type) {
	case int, int8, int16, int32, int64:
		// 整数类型，上下浮动指定比例，确保结果仍然是整数
		intVal := reflect.ValueOf(v).Int()
		variation := int64(float64(intVal) * variationRate)
		if variation == 0 {
			variation = 1 // 至少有1的变化
		}
		// 生成随机变化值，确保结果仍然是整数
		newVal := intVal + rand.Int63n(2*variation+1) - variation
		
		// 根据原始类型返回相应的整数类型
		switch v.(type) {
		case int:
			return int(newVal)
		case int8:
			return int8(newVal)
		case int16:
			return int16(newVal)
		case int32:
			return int32(newVal)
		case int64:
			return newVal
		default:
			return int(newVal)
		}

	case float32, float64:
		// 浮点数类型，上下浮动指定比例，保持原始精度
		floatVal := reflect.ValueOf(v).Float()
		variation := floatVal * variationRate
		newVal := floatVal + (rand.Float64()*2-1)*variation
		
		// 保持原始浮点数的精度
		origStr := fmt.Sprintf("%v", v)
		decimalPlaces := 0
		if dotIndex := strings.Index(origStr, "."); dotIndex != -1 {
			decimalPlaces = len(origStr) - dotIndex - 1
		}
		
		// 使用相同的精度格式化新值并返回相应的类型
		if decimalPlaces > 0 {
			roundedVal := math.Round(newVal*math.Pow10(decimalPlaces)) / math.Pow10(decimalPlaces)
			switch v.(type) {
			case float32:
				return float32(roundedVal)
			case float64:
				return roundedVal
			default:
				return roundedVal
			}
		}
		
		// 没有小数部分的情况
		switch v.(type) {
		case float32:
			return float32(newVal)
		case float64:
			return newVal
		default:
			return newVal
		}

	case string:
		// 检查是否是逗号分隔的数组（XML中的数组表示）
		if strings.Contains(v, ",") {
			// 处理逗号分隔的数组字符串
			parts := strings.Split(v, ",")
			for i := range parts {
				// 随机修改数组中的一些元素
				if rand.Float64() < 0.5 {
					parts[i] = randomizeString(parts[i])
				}
			}
			return strings.Join(parts, ",")
		} else if intVal, err := strconv.ParseInt(v, 10, 64); err == nil {
			// 是整数字符串
			variation := int64(float64(intVal) * variationRate)
			newVal := intVal + rand.Int63n(2*variation+1) - variation
			return strconv.FormatInt(newVal, 10)
		} else if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			// 是浮点数字符串
			variation := floatVal * variationRate
			newVal := floatVal + (rand.Float64()*2-1)*variation
			
			// 保持原始浮点数字符串的精度
			decimalPlaces := 0
			if dotIndex := strings.Index(v, "."); dotIndex != -1 {
				decimalPlaces = len(v) - dotIndex - 1
			}
			
			// 使用相同的精度格式化新值
			return strconv.FormatFloat(newVal, 'f', decimalPlaces, 64)
		} else {
			// 普通字符串，随机修改部分字符
			return randomizeString(v)
		}

	case bool:
		// 布尔值，有一定概率翻转
		if rand.Float64() < 0.5 {
			return !v
		}
		return v

	case []interface{}:
		// 数组，递归处理每个元素
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = generateVariation(item, variationRate)
		}
		return result

	case map[string]interface{}:
		// 对象，递归处理每个属性
		result := make(map[string]interface{})
		for key, item := range v {
			result[key] = generateVariation(item, variationRate)
		}
		return result

	default:
		// 其他类型，直接返回原值
		return v
	}
}

// randomizeString 随机修改字符串
// 字符串长度可进行10%范围内的变化，内容50%的字符随机变更
func randomizeString(s string) string {
	// 如果字符串很短，至少保证最小长度为1
	originalLen := len(s)
	if originalLen == 0 {
		return "a" // 空字符串返回一个随机字符
	}

	// 计算新长度：在原长度的90%-110%范围内随机变化
	minLen := int(float64(originalLen) * 0.9)
	if minLen < 1 {
		minLen = 1
	}
	maxLen := int(float64(originalLen) * 1.1)
	if maxLen < minLen {
		maxLen = minLen
	}
	
	// 随机确定新长度
	newLen := minLen + rand.Intn(maxLen-minLen+1)
	
	// 将原字符串转换为字符数组
	runes := []rune(s)
	originalRunes := make([]rune, len(runes))
	copy(originalRunes, runes)
	
	// 字符集：字母和数字的组合
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	// 如果新长度与原长度不同，需要调整字符串长度
	if newLen != originalLen {
		if newLen > originalLen {
			// 需要扩展字符串，在随机位置插入随机字符
			for i := 0; i < newLen-originalLen; i++ {
				insertPos := rand.Intn(len(runes) + 1)
				newChar := rune(charset[rand.Intn(len(charset))])
				// 在指定位置插入字符
				runes = append(runes[:insertPos], append([]rune{newChar}, runes[insertPos:]...)...)
			}
		} else {
			// 需要缩短字符串，随机删除字符
			for i := 0; i < originalLen-newLen; i++ {
				if len(runes) > 1 {
					deletePos := rand.Intn(len(runes))
					runes = append(runes[:deletePos], runes[deletePos+1:]...)
				}
			}
		}
	}
	
	// 随机变更50%的字符
	changeCount := len(runes) / 2
	if changeCount == 0 && len(runes) > 0 {
		changeCount = 1 // 至少变更一个字符
	}
	
	// 创建一个位置索引数组，用于随机选择要变更的位置
	positions := make([]int, len(runes))
	for i := range positions {
		positions[i] = i
	}
	
	// 随机打乱位置数组
	for i := len(positions) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		positions[i], positions[j] = positions[j], positions[i]
	}
	
	// 变更前changeCount个位置的字符
	for i := 0; i < changeCount && i < len(positions); i++ {
		pos := positions[i]
		runes[pos] = rune(charset[rand.Intn(len(charset))])
	}

	return string(runes)
}

// ConvertToXMLRows 将测试用例转换为XML行格式（每行一个完整的XML）
func ConvertToXMLRows(testCases []map[string]interface{}) [][]string {
	if len(testCases) == 0 {
		return [][]string{}
	}

	// 创建CSV数据，第一行是表头（只有一列：XML）
	result := make([][]string, 0, len(testCases)+1)
	// 添加表头
	result = append(result, []string{"XML"})

	// 填充数据行，每行包含一个完整的XML
	for _, testCase := range testCases {
		// 将测试用例转换为XML字符串
		xmlData, err := convertMapToXML(testCase)
		if err != nil {
			// 如果转换失败，使用JSON作为备选
			jsonData, _ := json.Marshal(testCase)
			xmlData = string(jsonData)
		}
		
		// 添加数据行（只有一列）
		result = append(result, []string{xmlData})
	}

	return result
}

// convertMapToXML 将map转换为XML字符串
func convertMapToXML(data map[string]interface{}) (string, error) {
	var xmlBuilder strings.Builder
	xmlBuilder.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	
	// 使用根元素包装所有字段
	xmlBuilder.WriteString("<root>\n")
	
	// 使用保存的原始字段顺序
	keys := originalKeyOrder
	if len(keys) == 0 {
		// 如果没有保存的顺序，则使用map的键
		keys = make([]string, 0, len(data))
		for key := range data {
			keys = append(keys, key)
		}
	}
	
	for _, key := range keys {
		value, exists := data[key]
		if !exists {
			continue
		}
		
		// 清理XML标签名（移除特殊字符）
		cleanKey := strings.ReplaceAll(key, " ", "_")
		cleanKey = strings.ReplaceAll(cleanKey, "-", "_")
		
		xmlBuilder.WriteString(fmt.Sprintf("  <%s>", cleanKey))
		
		// 根据值的类型进行处理
		switch v := value.(type) {
		case string:
			// 转义XML特殊字符
			escapedValue := escapeXMLValue(v)
			xmlBuilder.WriteString(escapedValue)
		case map[string]interface{}:
			// 嵌套对象，递归处理
			xmlBuilder.WriteString("\n")
			for nestedKey, nestedValue := range v {
				cleanNestedKey := strings.ReplaceAll(nestedKey, " ", "_")
				cleanNestedKey = strings.ReplaceAll(cleanNestedKey, "-", "_")
				xmlBuilder.WriteString(fmt.Sprintf("    <%s>%v</%s>\n", cleanNestedKey, nestedValue, cleanNestedKey))
			}
			xmlBuilder.WriteString("  ")
		case []interface{}:
			// 数组，处理每个元素
			xmlBuilder.WriteString("\n")
			for _, item := range v {
				xmlBuilder.WriteString(fmt.Sprintf("    <item>%v</item>\n", item))
			}
			xmlBuilder.WriteString("  ")
		default:
			// 其他类型直接转换为字符串
			xmlBuilder.WriteString(fmt.Sprintf("%v", v))
		}
		
		xmlBuilder.WriteString(fmt.Sprintf("</%s>\n", cleanKey))
	}
	
	xmlBuilder.WriteString("</root>")
	return xmlBuilder.String(), nil
}

// escapeXMLValue 转义XML特殊字符
func escapeXMLValue(value string) string {
	escaped := strings.ReplaceAll(value, "&", "&amp;")
	escaped = strings.ReplaceAll(escaped, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")
	escaped = strings.ReplaceAll(escaped, "\"", "&quot;")
	escaped = strings.ReplaceAll(escaped, "'", "&apos;")
	return escaped
}

// ConvertToJSONRows 将测试用例转换为单列JSON格式的CSV
// 每行包含一个完整的JSON字符串
func ConvertToJSONRows(testCases []map[string]interface{}) [][]string {
	if len(testCases) == 0 {
		return [][]string{}
	}

	// 创建CSV数据，第一行是表头
	result := make([][]string, 0, len(testCases)+1)
	// 添加表头（单列：JSON）
	result = append(result, []string{"JSON"})

	// 填充数据行
	for _, testCase := range testCases {
		// 将测试用例转换为JSON字符串
		jsonData, err := json.Marshal(testCase)
		if err != nil {
			// 如果转换失败，使用空JSON对象
			jsonData = []byte("{}")
		}
		
		// 添加数据行（只有一列）
		result = append(result, []string{string(jsonData)})
	}

	return result
}

// ConvertToCSV 将测试用例转换为CSV格式
func ConvertToCSV(testCases []map[string]interface{}) [][]string {
	if len(testCases) == 0 {
		return [][]string{}
	}

	// 使用保存的原始字段顺序
	keys := originalKeyOrder
	if len(keys) == 0 {
		// 如果没有保存的顺序，则使用第一个测试用例的键
		keys = make([]string, 0, len(testCases[0]))
		for key := range testCases[0] {
			keys = append(keys, key)
		}
	}

	// 创建CSV数据，第一行是表头
	result := make([][]string, 0, len(testCases)+1)
	// 添加表头
	result = append(result, keys)

	// 填充数据行
	for _, testCase := range testCases {
		row := make([]string, len(keys))
		for j, key := range keys {
			// 获取值
			value := testCase[key]
			
			// 根据原始数据类型处理值
			if originalValueTypes != nil {
				if typeInfo, ok := originalValueTypes[key]; ok {
					// 根据类型信息处理值
					if strings.HasPrefix(typeInfo, "int") {
						// 整数类型，确保输出为整数
						switch v := value.(type) {
						case float64:
							// 将浮点数转换为整数（JSON解析可能将整数解析为float64）
							row[j] = fmt.Sprintf("%.0f", v)
							continue
						case float32:
							// 将浮点数转换为整数
							row[j] = fmt.Sprintf("%.0f", v)
							continue
						case int, int8, int16, int32, int64:
							// 已经是整数
							row[j] = fmt.Sprintf("%d", v)
							continue
						}
					} else if strings.HasPrefix(typeInfo, "float") {
						// 浮点数类型，保持精度
						parts := strings.Split(typeInfo, ":")
						precision := 2 // 默认精度
						if len(parts) > 1 {
							if p, err := strconv.Atoi(parts[1]); err == nil {
								precision = p
							}
						}
						
						switch v := value.(type) {
						case float64, float32:
							floatVal := reflect.ValueOf(v).Float()
							row[j] = strconv.FormatFloat(floatVal, 'f', precision, 64)
							continue
						}
					} else if typeInfo == "array-string" {
						// XML中的数组（逗号分隔的字符串）
						if strVal, ok := value.(string); ok {
							// 直接使用字符串值
							row[j] = strVal
							continue
						}
					}
				}
			}
			
			// 默认处理方式
			switch v := value.(type) {
			case map[string]interface{}, []interface{}:
				// 将复杂结构转换为JSON
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					row[j] = fmt.Sprintf("%v", v)
				} else {
					row[j] = string(jsonBytes)
				}
			default:
				row[j] = fmt.Sprintf("%v", v)
			}
		}
		// 添加数据行
		result = append(result, row)
	}

	return result
}

// FormatJSON 格式化JSON字符串
func FormatJSON(jsonStr string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return "", err
	}

	prettyJSON, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", err
	}

	return string(prettyJSON), nil
}

// FormatXML 格式化XML字符串
func FormatXML(xmlStr string) (string, error) {
	// XML格式化比较复杂，这里简单实现
	// 实际项目中可能需要更复杂的处理
	xmlStr = strings.ReplaceAll(xmlStr, "><", ">\n<")
	return xmlStr, nil
}