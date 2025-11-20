# API 自动化测试命令行工具需求规格说明书

## 1. 项目概述

本项目旨在实现一个API自动化测试命令行工具（ai-auto-test-cmd，简称atc），用于简化API测试流程，提高测试效率。该工具支持通过LLM API生成测试用例，也支持本地生成测试用例，并能批量执行测试请求。目前已完整实现LLM API集成功能（llm-gen命令）、本地测试用例生成功能（local-gen命令）、智能约束系统、以及批量测试执行功能（request命令）。

## 2. 功能需求

### 2.1 测试用例生成功能

#### 2.1.1 通过LLM API生成测试用例

可通过命令和参数访问LLM API（访问的URL和API Key可通过命令行或配置文件config.toml提供），根据提供的XML或者JSON报文作为正向测试用例（即能被API正确处理的报文），生成多个API测试用例并保存为本地CSV文件，目前LLM后端提示词风格为生成边界情况的测试用例，用户也可通过命令行参数的形式附加自定义提示词，对生成的测试用例进行约束，支持通过`llm-gen`命令调用LLM API生成测试用例。

**可附加的参数包括：**
- 目标URL：指定Dify API的访问地址
- 请求参数格式：支持XML或JSON格式
- 请求参数：提供正例报文（支持命令行输入或文件输入）
- 生成用例数量：指定需要生成的测试用例数量
- 自定义提示词：提供附加的提示词以对生成的接口报文进行约束
  
**交互要求：**
- 执行时间可能较长，需提供友好的进度提示
- 可能出现失败情况，需提供清晰的错误信息
- 支持直接返回或文档方式返回生成的测试用例

**技术实现特性：**
- **流式响应处理**：支持Dify API的流式响应，实时处理返回数据
- **事件类型处理**：处理多种Dify事件类型（chatflow_started、chatflow_finished、text_chunk等）
- **错误处理机制**：当API连接失败或未返回有效数据时，提供详细错误信息
- **调试模式**：支持详细的调试输出，显示API请求和响应的完整信息
  - 通过 `--debug` 或 `-d` 参数启用调试模式
  - **HTTP请求详情显示**：
    - 完整的请求URL和HTTP方法（POST）
    - 详细的请求头信息（Authorization、Content-Type、Accept）
    - 格式化的JSON请求体内容，便于阅读
    - API密钥安全遮盖（只显示前4位和后4位，中间用星号遮盖）
  - **HTTP响应详情显示**：
    - HTTP响应状态码和状态描述
    - 完整的响应头信息（Content-Type、Transfer-Encoding等）
  - **流式响应调试**：
    - 实时显示Dify API流式响应的原始数据
    - 显示各种事件类型的详细信息（message、workflow_started、node_started、workflow_finished等）
    - 显示workflow_finished事件的完整数据结构
    - 显示节点执行状态、输入输出信息
    - 显示未处理的事件类型，便于调试和问题排查
  - **用户体验优化**：
    - 调试输出使用清晰的分隔线和emoji图标，便于阅读
    - 结构化的信息展示，方便开发者快速定位问题
    - 安全的敏感信息处理，保护API密钥等关键信息
- **参数验证**：严格验证必需参数（api-key、url），确保命令正确执行
- **动态用户标识**：基于用户当前IP地址生成唯一的用户标识，确保每次调用的唯一性和一致性
  - **IP地址获取**：
    - **主要方法**：通过连接外部地址（8.8.8.8:80）获取本机IP地址
    - **备用方法**：遍历网络接口获取非回环地址
    - **跨平台支持**：确保在Windows和Linux系统下都能正常工作
  - **IP地址处理**：
    - IPv4地址优先级高于IPv6地址
    - 将IP地址中的点号(.)和冒号(:)替换为下划线(_)
    - 生成格式：`ip_<IP地址>`（如：`ip_172_20_10_11`）
  - **容错机制**：如果IP获取失败，使用时间戳+随机字符串作为后备方案
    - 后备格式：`fallback_<时间戳><随机字符串>`
  - **安全考虑**：
    - 用户ID中不包含敏感信息
    - 仅使用本机IP地址，不涉及外部IP暴露
    - 生成的用户ID符合文件名安全规范
- **输出格式**：生成标准CSV格式文件，与local-gen命令保持一致的输出格式
- **Dify API参数配置**：
  - `post_type`：报文格式参数，传递json或xml格式标识
  - `test_num`：生成用例数量参数，指定需要生成的测试用例个数
  - `prompt`: 自定义提示词
- **智能解析系统**：支持多种测试用例格式的自动识别和解析
  - **JSON Array格式解析**：优先支持标准JSON数组格式的测试用例解析
    - 输入格式：`[{"data_field": value1}, {"data_field": value2}, ...]`
    - 自动识别并提取JSON数组中的每个对象作为独立测试用例
    - 支持从Markdown代码块中提取JSON数组内容
    - 支持复杂嵌套对象和多种数据类型（字符串、数字、对象等）
    - 内置JSON格式验证，确保生成的测试用例格式正确
    - 自动去重处理，避免重复的测试用例
  - **连续JSON对象格式解析**：支持解析多个连续的JSON对象（非数组格式）
    - 输入格式：`{"field1": value1} {"field2": value2} {"field3": value3}`
    - 自动识别和分割连续的JSON对象
    - 使用智能大括号匹配算法，正确处理嵌套结构
    - 支持字符串转义和特殊字符处理
    - 从Markdown代码块或纯文本中提取连续JSON对象
    - 自动跳过格式错误的JSON对象
  - **连续XML对象格式解析**：支持解析多个连续的XML对象
    - 输入格式：`<root>...</root>$$$$ <root>...</root>$$$$ <root>...</root>`
    - 自动识别和通过连续的4个$符号分割连续的XML对象
    - 使用智能标签匹配算法，正确处理嵌套结构和自闭合标签
    - 支持XML属性和复杂的标签结构
    - 从Markdown代码块或纯文本中提取连续XML对象
    - 自动跳过格式错误的XML对象
  - **智能格式检测**：
    - **JSON格式**：优先尝试JSON Array格式解析，失败时尝试连续JSON对象格式解析
    - **XML格式**：优先尝试连续XML对象格式解析
    - 支持混合格式文本的智能处理
    - 提供详细的解析状态和错误信息

**配置文件支持：**

llm-gen命令支持从配置文件读取URL和API Key，提供更便捷的使用方式：

- **默认配置文件**：`config.toml`
- **自定义配置文件**：通过`-c`参数指定
- **配置文件格式**：TOML格式
- **配置结构**：
  ```toml
  [llm]
  url = "http://localhost/v1"
  api_key = "app-uS9lBUYnv2Flm6mOylxxxxxx"
  ```
- **参数优先级**：命令行显式参数 > 配置文件参数
- **错误处理**：当配置文件不存在或解析失败时，提供详细错误信息
- **调试支持**：通过`-d`参数可显示配置文件读取过程

**自定义提示词支持：**

llm-gen命令支持多种方式使用自定义提示词，提供更灵活的测试用例生成控制：

- **配置文件方式**（推荐）：
  - 在配置文件的`[llm]`部分设置`user_prompt`字段
  - 支持多行文本，可包含详细的生成要求和指导
  - 与其他配置统一管理，便于维护
- **命令行文件方式**：
  - 通过`--prompt`参数指定提示词文件路径
  - 文件必须是UTF-8编码格式
  - 文件内容不能为空或仅包含空白字符
  - 建议不超过256个字符
- **优先级规则**：
  - 命令行指定的提示词文件优先级最高
  - 其次是配置文件中的`user_prompt`设置
  - 如果都未设置，则不使用自定义提示词
- **功能特性**：
  - 提示词内容作为`user_prompt`参数发送给Dify API
  - 支持调试模式显示提示词内容预览
  - 可与约束配置等其他功能结合使用
- **错误处理**：
  - 文件不存在时提供明确错误提示
  - UTF-8编码验证失败时显示详细错误信息
  - 空文件或仅包含空白字符时给出相应提示

**使用示例：**

```shell
# 根据正例xml报文生成5条测试用例（显式指定URL和API Key）
atc llm-gen --url https://xxx.llm.com/xxx/xxx --api-key your_api_key --xml "<root><name>test</name></root>" --num 5

# 根据正例json报文生成8条测试用例
atc llm-gen --url https://xxx.llm.com/xxx/xxx --api-key your_api_key --json '{"name":"test","age":25}' --num 8

# 使用配置文件中的正例报文生成测试用例
atc llm-gen -c config.toml --num 6

# 命令行参数覆盖配置文件中的正例报文
atc llm-gen -c config.toml --json '{"name":"test"}' --num 3

# 指定输出文件名生成测试用例
atc llm-gen --url https://xxx.llm.com/xxx/xxx --api-key your_api_key --json '{"test": "hello"}' --num 5 --output custom_test_cases.csv

# 启用调试模式，显示详细的HTTP请求和响应信息
atc llm-gen --url https://xxx.llm.com/xxx/xxx --api-key your_api_key --xml --file input.xml --num 7 --debug

# 启用调试模式的简写形式，显示完整的请求头、请求体和响应详情
atc llm-gen --url https://xxx.llm.com/xxx/xxx --api-key your_api_key --json '{"name":"test"}' --num 4 -d

# 使用默认配置文件(config.toml)中的URL和API Key
atc llm-gen --xml "<root><name>test</name></root>" --num 5

# 使用指定配置文件中的URL和API Key
atc llm-gen -c my-config.toml --xml "<root><name>test</name></root>" --num 5

# 从配置文件读取URL，命令行指定API Key（混合使用）
atc llm-gen -c config.toml --api-key your_api_key --json '{"name":"test"}' --num 3

# 使用配置文件并启用调试模式，显示配置文件读取过程
atc llm-gen -c custom-config.toml --xml "<root><name>test</name></root>" --num 5 --debug

# 使用配置文件中的自定义提示词生成测试用例
atc llm-gen --xml "<user><name>张三</name><age>25</age></user>" -c config.toml --num 3

# 使用命令行指定的提示词文件（优先级高于配置文件）
atc llm-gen -c my-config.toml --json '{"name":"test"}' --prompt prompt.txt --num 5

# 使用提示词文件并启用调试模式
atc llm-gen --xml "<test>data</test>" --prompt custom_prompt.txt --num 2 --debug
```

**JSON Array格式示例：**

当Dify API返回JSON Array格式的测试用例时，系统会自动识别并解析：

```json
[
  {
    "data_field": 12345
  },
  {
    "data_field": "test_user"
  },
  {
    "data_field": "test@example.com"
  },
  {
    "user_id": 67890,
    "username": "admin",
    "email": "admin@example.com"
  }
]
```

系统会将上述JSON数组中的每个对象解析为独立的测试用例，并保存到CSV文件中。

**连续JSON对象格式示例：**

当Dify API返回连续JSON对象格式的测试用例时，系统会自动识别并解析：

```json
{
  "name": "张三",
  "age": 25,
  "email": "zhangsan@example.com",
  "phone": "13812345678"
}
{
  "name": "李四",
  "age": 26,
  "email": "lisi@example.com",
  "phone": "13912345678"
}
{
  "name": "王五",
  "age": 27,
  "email": "wangwu@example.com",
  "phone": "13712345678"
}
```

系统会自动分割上述连续的JSON对象，将每个对象解析为独立的测试用例，并保存到CSV文件中。

**连续XML对象格式示例：**

当Dify API返回连续XML对象格式的测试用例时，系统会自动识别并解析：

```xml
<user>
  <name>张三</name>
  <age>25</age>
  <email>zhangsan@example.com</email>
  <phone>13812345678</phone>
</user>$$$$
<user>
  <name>李四</name>
  <age>26</age>
  <email>lisi@example.com</email>
  <phone>13912345678</phone>
</user>$$$$
<user>
  <name>王五</name>
  <age>27</age>
  <email>wangwu@example.com</email>
  <phone>13712345678</phone>
</user>$$$$
```

系统会根据每个XML结尾的4个$$$$分割上述连续的XML对象，将每个对象解析为独立的测试用例，并保存到CSV文件中。

#### 2.1.2 本地生成测试用例

可以在不调用Dify API的情况下随机生成测试用例，用户只需提供API正向用例XML或JSON报文，即可按照下述的规则改变输入正例报文每个字段的值，生成任意数量的测试用例报文。

**生成规则：**
- 根据正向用例值域上下浮动自动生成测试数据（默认浮动范围50%，可通过配置文件自定义）
  - 对于数字类型：值会在原始值的±浮动范围内随机变化
  - 对于字符串类型：长度可进行10%范围内的变化（90%-110%），内容中指定比例的字符变更为随机字符串（包含字母和数字）
  - 对于布尔类型：有指定概率会翻转值
  - 对于数组类型：每个元素都会按照上述规则进行随机变化
  - 对于嵌套对象：递归应用上述规则到每个字段
- 用户可以指定生成用例的数量
- 支持将生成的测试用例保存为CSV文件，便于后续使用
- 支持通过配置文件中的`testcase.variation_rate`参数自定义随机化程度（0.0-1.0，默认0.5）

**使用示例：**

```shell
# 本地根据正例xml报文生成10条测试用例（生成单列XML格式）
atc local-gen --xml "<user><name>张三</name><age>25</age></user>" -n 10

# 本地根据正例json报文生成15条测试用例（生成单列JSON格式）
atc local-gen --json '{"name":"张三","age":25}' -n 15

# 使用配置文件中的正例报文生成测试用例
atc local-gen -c config.toml -n 10

# 命令行参数覆盖配置文件中的正例报文
atc local-gen -c config.toml --json '{"name":"test"}' -n 15

# 本地根据正例json报文生成5条测试用例并保存到指定文件
atc local-gen --json '{"name":"张三","age":25}' -n 5 -o test_data.csv

# 使用配置文件中的约束配置生成智能测试用例
atc local-gen --json -n 10 -c config.toml

# 使用自定义配置文件生成测试用例
atc local-gen --json -n 10 -c my-config.toml

# 使用自定义随机化因子生成测试用例（配置文件中设置variation_rate = 0.3）
atc local-gen --json '{"name":"张三","age":25,"price":100.5}' -n 10 -c low-variation.toml
```

以上是在Unix系统下的使用示例，如果使用cmd或者Powershell，需要对从命令行输入的JSON字符串中的双引号进行转义，确保命令行解析正确，例：

```shell
# cmd 使用双引号，JSON中的双引号使用反斜线转义
atc local-gen --json "{\"name\":\"张三\",\"age\":25}" -n 5 -o test_data.csv

# Powershell 使用单引号，JSON中的双引号使用反斜线转义
atc local-gen --json '{\"name\":\"张三\",\"age\":25}' -n 5 -o test_data.csv
```

**数据处理能力：**
- 智能解析XML和JSON格式，支持复杂的嵌套结构
- 自动识别数据类型（数字、字符串、布尔值等）并应用相应的随机化规则
- 保留原始数据的结构和关系，只修改值
- **字段顺序一致性**：生成的测试用例字段顺序与正例输入保持完全一致
- **XML格式输出**：生成单列CSV文件，每行包含一个完整的XML文档，便于直接作为请求体使用
- **JSON格式输出**：生成单列CSV文件，每行包含一个完整的JSON字符串，便于直接作为请求体使用

**输入方式支持：**
- **命令行输入**：通过 `--xml` 或 `--json` 参数直接在命令行中提供正例报文（可选，用于覆盖自动检测）
- **自动格式检测**：`request` 命令支持从CSV文件第一行自动检测格式，无需手动指定参数
- **配置文件输入**：通过配置文件中的 `positive_example` 字段设置正例报文，支持多行字符串
- **实时格式验证**：输入的XML/JSON内容会立即进行格式验证，确保语法正确性
- **详细错误提示**：格式验证失败时提供详细的错误信息和位置定位

**使用示例：**
```bash
# 使用方式
atc llm-gen --xml "<root><name>test</name></root>" -n 5
atc local-gen --json '{"name":"test","age":25}' -n 10
```

**输入文件格式要求：**
- **XML文件格式**：
  - 必须是标准完整的XML格式，符合XML语法规范
  - 文件内容不能为空或仅包含空白字符
  - 必须具有正确的XML结构（开始标签、结束标签匹配等）
  - 支持复杂的嵌套结构和属性
- **JSON文件格式**：
  - 必须是标准完整的JSON格式，符合JSON语法规范
  - 文件内容不能为空或仅包含空白字符
  - 必须具有正确的JSON结构（括号匹配、引号正确等）
  - 支持复杂的嵌套对象和数组结构
- **格式验证机制**：
  - 系统会在读取文件后立即进行格式验证
  - 如果文件格式不符合要求，会显示详细的错误信息并终止执行
  - 命令行输入的报文同样会进行格式验证
  - 验证失败时会提供具体的错误位置和原因，便于用户修正

#### 2.1.3 智能约束系统

本地生成测试用例功能支持智能约束系统，能够根据字段名自动应用相应的数据生成约束，约束可以通过全局配置文件声明，使得生成内容不会过分随机化，能生成满足API测试需求、更加真实、符合业务场景的测试数据的大量用例。

**约束系统特性：**
- **智能字段识别**：根据字段名自动匹配相应的约束规则
- **多种数据类型支持**：日期、中文姓名、手机号、邮箱、地址、身份证号、数值等
- **统一配置管理**：约束配置集成在全局配置文件中，与Dify API配置统一管理
- **向后兼容**：不影响原有的随机变化生成模式

**使用示例：**

```shell
# 使用配置文件中的约束配置生成测试用例
atc local-gen -n 5 -c config.toml -o output.csv

# 使用自定义配置文件生成测试用例
atc local-gen -n 5 -c custom.toml -o output.csv

# 不使用约束，采用原有随机变化模式（不指定配置文件）
atc local-gen --json -n 5 -o output.csv

# 验证配置文件
atc validate

# 验证指定的配置文件
atc validate my-config.toml

# 验证配置文件并显示详细信息
atc validate --verbose config.toml
```

**约束系统开关：**

约束系统支持通过配置文件进行开关控制，提供灵活的使用方式：

- **开关配置**：`[constraints].enable`
  
- **开关行为**：
  - `true`：启用约束系统，使用智能约束模式生成测试数据
  - `false`：禁用约束系统，使用随机变化模式生成测试数据
  - 未设置：根据是否存在约束配置自动决定（有约束配置则启用，否则禁用）

**全局配置文件（config.toml）：**

约束系统通过全局配置文件定义字段约束规则，该配置文件同时包含Dify API配置、约束配置和内置数据。约束配置必须放在`[constraints]`根节点下，避免与其他配置项冲突。支持以下约束类型：

0. **保持原值约束（keep_original）**：
   - 支持字段：任意字段或节点
   - 功能描述：保持字段原始值不变的约束类型
   - 适用场景：当需要某些字段在所有测试用例中保持与原始输入相同的值时使用
   - 优先级规则：子字段的独立约束优先于父节点的keep_original约束
   - 生成效果：字段值与原始输入完全一致，不进行任何变化

1. **日期字段约束**：
   - 支持字段：`date`、`create_date`、`update_date`等
   - 配置项：日期格式、最小日期、最大日期
   - 生成效果：20230101（YYYYMMDD格式）

2. **中文姓名约束**：
   - 支持字段：`name`、`user_name`、`username`、`real_name`等
   - 内置中文姓名数据集（姓氏+名字组合）
   - 生成效果：张伟、王芳、李明等真实中文姓名

3. **年龄约束**：
   - 支持字段：`age`
   - 配置项：最小值、最大值（1-120）
   - 生成效果：符合实际年龄范围的整数

4. **手机号约束**：
   - 支持字段：`phone`、`mobile`、`phone_number`等
   - 生成规则：中国大陆手机号格式（1[3-9]xxxxxxxxx）
   - 生成效果：13812345678、18987654321等

5. **邮箱约束**：
   - 支持字段：`email`、`email_address`等
   - 内置常见邮箱域名（qq.com、163.com、gmail.com等）
   - 生成效果：user123@qq.com、demo456@163.com等

6. **价格约束**：
   - 支持字段：`price`、`amount`、`money`等
   - 配置项：最小值、最大值、精度（小数位数）
   - 生成效果：123.45、9999.99等符合价格格式的浮点数

7. **中文地址约束**：
   - 支持字段：`address`、`addr`、`location`等
   - 内置中国主要城市地址数据集
   - 生成效果：北京市朝阳区建国门外大街1号等真实地址

8. **身份证号约束**：
   - 支持字段：`id_card`、`identity_card`、`card_no`等
   - 生成规则：符合中国身份证号格式（18位）
   - 生成效果：110101199001011234等

9. **数量约束**：
   - 支持字段：`quantity`、`count`等
   - 配置项：最小值、最大值
   - 生成效果：符合业务逻辑的数量值

10. **状态码约束**：
    - 支持字段：`status`
    - 配置项：状态码范围（0-9）
    - 生成效果：符合系统状态定义的整数值

11. **ID编号约束**：
    - 支持字段：`id`、`user_id`、`order_id`等
    - 配置项：ID范围（1-999999）
    - 生成效果：符合系统ID规范的整数值

12. **日期时间约束（datetime）**：
    - 支持字段：`datetime`、`create_datetime`、`update_datetime`、`last_login_time`、`txn_dt`等
    - 功能描述：生成符合RFC 3339 Extended标准的日期时间值
    - 配置项：
      - `min_datetime`：最小日期时间（可选，默认为2020-01-01T00:00:00Z）
      - `max_datetime`：最大日期时间（可选，默认为2030-12-31T23:59:59Z）
      - `timezone`：时区设置（可选，支持UTC、+08:00等格式）
    - 生成效果：
      - UTC时区：`2024-11-22T15:00:26.431Z`
      - 带偏移量时区：`2024-06-29T07:10:19.531+08:00`
      - 毫秒精度支持，完全符合RFC 3339 Extended标准
    - 技术特性：
      - 支持嵌套字段名匹配（如`Body.CreateDatetime`自动匹配配置中的`CreateDatetime`）
      - 智能字段名匹配（支持大小写不敏感匹配）
      - 生成的时间值可被Python `datetime.fromisoformat()` 正确解析
      - 支持多种时区格式输出

13. **银行卡号约束（bank_card）**：
    - 支持字段：`bank_card`、`bank_card_number`、`card_number`、`account_number`等
    - 功能描述：从内置银行卡号数据集中随机选择银行卡号
    - 生成效果：6222021234567890、6227001234567896等真实银行卡号格式
    - 内置银行卡号数据集
    - 技术特性：
      - 支持自定义银行卡号数据集
      - 从预定义的银行卡号列表中随机选择
      - 确保生成的卡号格式正确

**内置数据集：**

约束系统包含丰富的内置中文数据集：
- **中文姓名数据集**：包含常见的中文姓氏和名字，支持真实姓名组合
- **中文地址数据集**：包含中国主要城市的真实地址信息
- **邮箱域名数据集**：包含常见的邮箱服务提供商域名
- **银行卡号数据集**：包含中国主要银行的完整银行卡号示例，支持生成真实格式的银行卡号

**约束匹配规则：**
- 优先进行精确字段名匹配
- 支持小写字段名匹配
- 未匹配到约束的字段将使用原有的随机变化逻辑
- 约束配置支持热更新，无需重新编译程序

**配置文件示例：**

```toml
# API自动化测试命令行工具配置文件
# 包含LLM API配置、请求配置、用例设置、约束配置和内置数据

[llm]
# LLM API Base URL
url = "http://localhost/v1"

# LLM API Key
api_key = "app-uS9lBUYnv2Flm6mOylxhggy7"

# 自定义提示词（可选）
user_prompt = "请生成边界情况的测试用例，包括空值、极值、特殊字符等场景"

[request]
# 目标URL
url = "https://httpbin.org/post"
# 请求方法
method = "post"
# 请求超时时间（秒）
timeout = 30
# 并发请求数
concurrent = 3

[testcase]
# 用例生成数量
num = 10
# 输出文件路径
output = "test_cases.csv"
# 正例报文类型（xml或json）
type = "json"
# 随机化因子，控制数据变化程度（0.0-1.0，默认0.5）
variation_rate = 0.5

# 约束系统配置
[constraints]
# 约束系统开关（true: 启用约束系统，false: 使用随机变化模式）
enable = true

# 保持原值约束示例
[constraints.Head]
type = "keep_original"
description = "保持Head节点所有字段原值不变"

# 子字段独立约束（优先级高于父节点的keep_original）
[constraints.Head.TxnDt]
type = "date"
start_date = "2020-01-01"
end_date = "2030-12-31"
format = "yyyyMMdd"
description = "交易日期字段，能覆盖父节点的keep_original约束"

# 日期字段约束
[constraints.date]
type = "date"
format = "20060102"  # Go时间格式
min_date = "20200101"
max_date = "20301231"
description = "日期字段，格式为YYYYMMDD"

# 姓名字段约束
[constraints.name]
type = "chinese_name"
description = "中文姓名"

# 年龄字段约束
[constraints.age]
type = "integer"
min = 1
max = 120
description = "年龄范围1-120"

# 日期时间字段约束
[constraints.CreateDatetime]
type = "datetime"
min_datetime = "2020-01-01T00:00:00Z"
max_datetime = "2025-12-31T23:59:59Z"
timezone = "UTC"
description = "创建时间字段，UTC时区"

[constraints.UpdateDatetime]
type = "datetime"
min_datetime = "2024-01-01T00:00:00+08:00"
max_datetime = "2024-12-31T23:59:59+08:00"
timezone = "+08:00"
description = "更新时间字段，东八区时区"

[constraints.TxnDt]
type = "datetime"
min_datetime = "2024-01-01T00:00:00+08:00"
max_datetime = "2030-12-31T23:59:59+08:00"
timezone = "+08:00"
description = "交易日期时间字段"

# 内置数据集
[constraints.builtin_data]
first_names = ["张", "王", "李", "赵", "刘"]
last_names = ["伟", "芳", "娜", "敏", "静"]
addresses = ["北京市朝阳区建国门外大街1号", "上海市浦东新区陆家嘴环路1000号"]
email_domains = ["qq.com", "163.com", "126.com", "gmail.com"]
bank_cards = ["6222021234567890", "6227001234567896", "6228481234567893", "6217851234567899", "6225881234567892"]
```

**自定义类型与数据集：**

在 `[constraints.types]` 下声明可复用的自定义类型；类型可通过内联 `values`、引用自定义 `datasets` 或内置数据集，并可选用正则 `pattern` 对候选值进行过滤。自定义数据集在 `[constraints.datasets]` 下声明，字段通过 `type` 引用相应类型。

```toml
[constraints]
enable = true

# 声明自定义类型（支持 values / dataset / pattern）
[constraints.types.status_text]
values = ["pending", "paid", "shipped", "cancelled"]
description = "订单状态文本"

[constraints.types.region_code]
dataset = "region_codes"           # 引用下方的自定义数据集
pattern = "^[A-Z]{2}-\\d{2}$"     # 使用正则过滤候选值
description = "区域编码"

# 引用内置数据集作为自定义类型的数据源
[constraints.types.email_domain_custom]
dataset = "email_domains"           # 来自内置数据集
description = "邮箱域名（内置数据集）"

# 自定义数据集（供自定义类型引用）
[constraints.datasets]
region_codes = ["CN-01", "CN-02", "US-01", "US-02", "JP-01", "DE-01"]

# 在字段上应用自定义类型
[constraints.order.status_text]
type = "status_text"

[constraints.user.region]
type = "region_code"

[constraints.user.email_domain]
type = "email_domain_custom"
```

注意事项：
- `atc validate` 会检查 `pattern` 是否可编译，实际匹配在生成阶段执行；无效正则会被安全忽略。
- 未知的 `dataset` 会被跳过；若同时定义了 `values` 则回退使用 `values`。
- 当 `values` 和 `dataset` 都为空时，字段保持原值不变。
- 使用 `atc validate -v` 可查看自定义类型和自定义数据集的统计信息。

**嵌套字段约束：**

约束解析支持嵌套表路径，系统会递归收集叶子字段约束。顶层对象如 `[constraints.user]`、`[constraints.order]` 不会被误识别为具体字段约束，只有叶子节点（如 `[constraints.user.profile.status_text]`）会被视为字段约束并应用到生成阶段。字段名匹配规则同时支持 `.` 和 `[]` 形式的层级映射（例如 `user.profile[0].status_text`）。

```toml
[constraints.user.profile.status_text]
type = "status_text"

[constraints.order.status_text]
type = "status_text"
```

**组合约束（Composite Constraints）：**

当多个字段之间存在强相关性（必须成对/成组出现）时，可使用组合约束确保这些字段在生成阶段共同取值。例如 `province` 与 `city` 必须来自同一地区映射。

语法示例：

```toml
[constraints]
enable = true

# 多字段相关性示例：省份与城市需要成组选择
[constraints.composite_address]
type = "composite"
fields = ["province", "city"]
data = [["上海", "上海市"], ["北京", "北京市"], ["广东", "广州市"]]
description = "组合约束：province 与 city 一组选择，保持一致性"
```

生成行为：
- 每个测试用例会对每个组合约束随机选择一行 `data`（组级选择）。
- 该组合中的所有字段（匹配依据为“路径最后一段的字段名”）将被覆盖为所选行中的对应值。
- 支持嵌套对象与数组内对象：匹配发生在叶子字段（例如 `user.address.province`、`addresses[0].city`）。
- 字段名匹配采用归一化规则：名称转换为小写、`-` 转为 `_`（例如 `city-name` 与 `city_name` 视为一致）。

验证规则：
- `fields` 不可为空且必须唯一；`data` 不可为空。
- `data` 中每一行的元素数量必须与 `fields` 数量一致。
- 若配置非法，`atc validate` 将报告详细错误信息。

当前限制与注意事项：
- 同一测试用例内，组合字段的所有出现点将使用同一组值（一次选择，处处覆盖）。如数组中有多个对象，其对应字段会一致，以保证一致性。
- 组合约束仅对基础类型字段覆盖生效；若父级字段设置了 `keep_original`，子字段仍会按组合约束覆盖。

**生成效果对比：**

使用约束系统前（随机变化）：
```json
{"date":"27388202","name":"p","age":18,"phone":"11684695289","email":"1haDgsai8xOmpyU.C0m","price":122,"address":"K京w区","id_card":"71614056398366167","create_datetime":"20a4-0EY-15TADy00s0nxJ005LQM:QU","update_datetime":"6024-0uiiPb14:i60200kz00M"}
```

使用约束系统后（智能约束）：
```json
{"date":"20230101","name":"周桂兰","age":64,"phone":"17234495798","email":"test473@189.cn","price":161782.59,"address":"武汉市武昌区中南路99号","id_card":"500101198909148195","bank_card":"6222021234567890","create_datetime":"2024-11-22T15:00:26.431Z","update_datetime":"2024-06-29T07:10:19.531+08:00"}
```

#### 2.1.4 配置文件验证

系统提供了配置文件验证功能，确保配置文件的正确性和完整性。

**验证功能特性：**

1. **格式验证**
   - Dify API配置有效性检查
   - 约束类型有效性检查
   - 必需字段完整性验证
   - 数据类型正确性验证

2. **内容验证**
   - 日期格式和范围验证
   - 数值范围合理性检查
   - 精度设置有效性验证
   - 内置数据完整性检查

3. **错误报告**
   - 详细的错误信息提示
   - 多错误批量报告
   - 错误位置精确定位

**验证命令：**

```shell
# 验证默认配置文件
atc validate

# 验证指定配置文件
atc validate my-config.toml

# 显示详细验证信息
atc validate --verbose
```

**验证规则：**

- **Dify API配置验证**：检查URL和API Key配置的完整性
- **约束类型验证**：支持的类型包括 `keep_original`、`date`、`chinese_name`、`phone`、`email`、`chinese_address`、`id_card`、`bank_card`、`integer`、`float`、`datetime`
- **keep_original约束验证**：该约束类型不需要额外参数，主要验证配置语法正确性
- **日期约束验证**：日期格式必须符合Go时间格式规范，最小日期不能晚于最大日期
- **数值约束验证**：最大值不能小于最小值，整数类型不应设置精度字段
- **内置数据验证**：姓氏、名字、地址列表不能为空，邮箱域名格式验证，银行卡号格式和长度验证
- **自定义类型与数据集验证**：`types` 支持 `values`、`dataset` 与 `pattern`；`pattern` 仅进行编译检查，实际匹配在生成阶段；未知 `dataset` 将被安全忽略并回退到 `values`；空类型声明不会导致错误，字段将保留原值；`atc validate -v` 输出自定义类型与数据集统计信息。

**验证示例：**

正确配置验证：
```
🔍 正在验证配置文件: config.toml
✅ 配置文件验证通过！

📊 配置文件统计信息:
  • Dify API配置:
    - URL: http://localhost/v1
    - API Key: app-****ggy7
    - 自定义提示词: 请生成边界情况的测试用例，包括空值、极值、特殊字符等场景
  • 约束字段总数: 30
  • 约束类型分布:
    - keep_original: 1 个
    - chinese_name: 4 个
    - date: 3 个
    - integer: 8 个
    - float: 3 个
    - phone: 3 个
    - email: 2 个
    - chinese_address: 3 个
    - id_card: 3 个
    - bank_card: 4 个
  • 内置数据集:
    - 姓氏: 20 个
    - 名字: 52 个
    - 地址: 15 个
    - 邮箱域名: 10 个
    - 银行卡号: 15 个
```

错误配置验证：
```
🔍 正在验证配置文件: invalid.toml
❌ 验证失败:
约束配置验证失败: 发现 3 个配置错误:
字段 'invalid_field': 无效的约束类型 'invalid_type'，支持的类型: date, chinese_name, phone, email, chinese_address, id_card, integer, float
字段 'missing_type': 约束类型 'type' 不能为空
字段 'bad_range': 最大值 50.00 不能小于最小值 100.00
```

### 2.2 测试执行功能

可通过命令及本地的CSV文件，批量请求目标系统接口，返回执行结果，结果默认自动保存，目前开发者需根据执行结果自行判断是否通过测试。

**功能特点：**
- 支持POST和GET等HTTP方法
- 支持直接在命令行显示结果
- 结果默认自动保存到指定文件
- **配置文件支持**：支持从配置文件读取所有参数，简化命令行使用
- **鉴权支持**：支持多种鉴权方式（Bearer Token、Basic Auth、API Key等）
- **自定义HTTP头**：支持添加自定义HTTP请求头，满足不同接口的要求
- **参数优先级**：命令行参数优先级高于配置文件参数

**使用示例：**

```shell
# 自动检测CSV文件格式，无需指定--json或--xml参数
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv

# 手动指定JSON格式（覆盖自动检测结果）
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --json

# 手动指定XML格式（覆盖自动检测结果）
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --xml

# 根据测试用例文件xxx.csv,批量使用GET方法请求目标系统http接口，数据放在请求体中
atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv

# 添加URL查询参数（适用于任何请求方法）
atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv --query "version=v1" --query "debug=true"

# 根据测试用例文件xxx.csv,批量使用GET方法请求目标系统http接口，结果保存至指定目录及文件
atc request -u https://xxx.system.com/xxx/xxx -m get -f xxx.csv /xxx/tool/result.csv

# 启用调试模式，详细输出每个请求的URL、HTTP头和请求体信息，以及响应详情
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --debug

# 使用配置文件中的参数（支持自动格式检测）
atc request -c config.toml

# 使用配置文件，命令行参数覆盖配置文件中的设置
atc request -c config.toml -u https://api.example.com/test

# 使用Bearer Token鉴权发送请求
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --auth-bearer "your_token_here"

# 使用Basic Auth鉴权发送请求
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --auth-basic "username:password"

# 添加自定义HTTP头发送请求
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --header "X-API-Key: your_api_key" --header "X-Client-Version: 1.0"

# 组合使用鉴权和自定义头
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --auth-bearer "token" --header "X-Request-ID: 12345"

# 组合使用查询参数、鉴权和自定义头
atc request -u https://xxx.system.com/xxx/xxx -m post -f xxx.csv --query "api_version=2.0" --auth-bearer "token" --header "X-Request-ID: 12345"

# 忽略TLS证书验证错误（适用于自签名证书或测试环境）
atc request -u https://self-signed.example.com/api -m post -f xxx.csv --ignore-tls

# 在配置文件中设置忽略TLS证书验证
atc request -c config.toml  # 配置文件中设置 ignore_tls_errors = true
```

**配置文件支持：**

request命令支持从配置文件读取所有参数，配置文件格式如下：

```toml
[request]
url = "https://api.example.com/test"
method = "post"
file = "test_cases.csv"
save_path = "results.csv"
timeout = 30
concurrent = 3
ignore_tls_errors = false  # 是否忽略TLS证书验证错误，适用于自签名证书或测试环境
auth_bearer = "your_bearer_token_here"
auth_basic = "username:password"
auth_api_key = "your_api_key_here"
headers = [
    "X-API-Version: v1",
    "X-Client-ID: atc-tool",
    "X-Request-Source: automated-test"
]
query = [
    "version=v1",
    "source=atc-tool",
    "debug=true"
]
```

### 2.3 一键生成并执行功能

**功能概述：**
一键生成并执行功能是ATC工具的核心特性之一，通过在测试用例生成命令中添加`--exec`（或`-e`）参数，实现测试用例生成和接口测试的无缝集成。该功能消除了传统的"生成→保存→执行"三步流程，提供了更高效的测试体验。

**支持命令：**
- `local-gen --exec`：本地生成测试用例并立即执行
- `llm-gen --exec`：通过LLM API生成测试用例并立即执行

**核心特性：**

1. **参数复用机制**：
   - 自动复用生成命令中的所有请求参数（URL、方法、格式、鉴权等）
   - 无需重复输入相同的配置信息
   - 保持生成和执行阶段的参数一致性

2. **内存直接传递**：
   - 生成的测试用例直接在内存中传递给执行引擎
   - 避免临时文件的创建和读取，提高执行效率
   - 减少磁盘I/O操作，提升性能

3. **完整功能支持**：
   - 支持所有`request`命令的功能特性
   - 包括鉴权、自定义HTTP头、调试模式、结果保存等
   - 与独立的`request`命令功能完全一致

4. **向后兼容性**：
   - 不影响原有的独立生成和执行流程
   - `--exec`参数为可选参数，默认行为保持不变
   - 现有脚本和工作流程无需修改

**使用示例：**

```shell
# 本地生成JSON格式测试用例并立即执行POST请求
atc local-gen --json '{"name":"张三","age":25}' -n 5 \
  --exec -u https://api.example.com/users -m post

# 本地生成XML格式测试用例并立即执行，使用Bearer Token鉴权，设置3并发
atc local-gen --xml -C 3 \
  --exec -u https://api.example.com/data -m post \
  --auth-bearer "your_token_here"

# 通过LLM API生成测试用例并立即执行（request参数从配置文件读取）
atc llm-gen --json '{"user":"admin","action":"login"}' -n 10 \
  -c config.toml --exec --debug

# 生成测试用例并执行，结果自动保存到文件
atc local-gen --json '{"product":"laptop","price":5000}' -n 8 \
  --exec -u https://api.example.com/products -m post \
  /path/to/results.csv

# 使用约束系统生成测试用例并立即执行（request参数从配置文件读取）
atc local-gen --json '{"name":"","phone":"","email":""}' -n 5 \
  -c config.toml --exec

# 组合使用多种功能：约束、用例设置、一键执行（所有参数从配置文件读取）
atc local-gen -c user_config.toml --exec
```

**技术实现特性：**

1. **统一参数验证**：
   - 在执行阶段复用生成阶段已验证的参数
   - 避免重复验证，提高执行效率
   - 确保参数的一致性和正确性

2. **错误处理机制**：
   - 生成阶段失败时，不会进入执行阶段
   - 执行阶段的错误处理与独立`request`命令一致
   - 提供清晰的错误信息和调试支持

3. **性能优化**：
   - 内存中直接传递测试用例数据
   - 避免文件系统的读写操作
   - 减少数据序列化和反序列化的开销

4. **调试支持**：
   - 支持`--debug`参数，详细输出生成和执行过程
   - 显示生成的测试用例数量和内容
   - 提供完整的请求和响应信息

**工作流程：**

1. **参数解析**：解析命令行参数，识别`--exec`标志
2. **测试用例生成**：根据指定的生成方式（本地或Dify API）生成测试用例
3. **参数传递**：将请求相关参数传递给执行引擎
4. **直接执行**：在内存中直接执行生成的测试用例
5. **结果输出**：显示执行结果和统计信息

**优势对比：**

| 传统流程 | 一键执行流程 |
|---------|-------------|
| 1. 生成测试用例 | 1. 生成并执行测试用例 |
| 2. 保存到CSV文件 | （无需中间步骤） |
| 3. 读取CSV文件执行 | |
| **3个步骤，需要文件I/O** | **1个步骤，纯内存操作** |

**适用场景：**
- 快速API测试和验证
- 持续集成/持续部署(CI/CD)流程
- 开发阶段的接口调试
- 自动化测试脚本
- 性能测试和压力测试

**实现特性：**
- 支持从CSV文件读取测试数据，自动解析为测试用例
- 支持POST和GET请求方法
- 支持多种请求体格式：JSON和XML（支持自动检测或手动指定格式）
- **自动格式检测**：从CSV文件第一行自动识别请求体格式（XML或JSON），无需手动指定
- **手动格式指定**：可选使用 `--json` 或 `--xml` 参数强制指定请求体格式，覆盖自动检测结果
- **智能CSV格式识别**：
  - **自动格式检测**：从CSV文件第一行自动识别请求体格式
    - 当第一行包含 "xml"（不区分大小写）时，自动识别为XML格式
    - 当第一行包含 "json"（不区分大小写）时，自动识别为JSON格式
    - 无需手动指定 `--xml` 或 `--json` 参数
  - **手动格式覆盖**：使用 `--xml` 或 `--json` 参数可以覆盖自动检测结果
  - **格式处理逻辑**：
    - XML格式：自动识别单列XML格式（列名为"XML"），直接使用每行的XML内容作为请求体
    - JSON格式：自动识别单列JSON格式（列名为"JSON"），直接使用每行的JSON内容作为请求体；对于多列格式，将各列数据组合为JSON请求体
- 支持并发请求以提高执行效率
- 支持自定义超时时间
- 智能数据类型解析（数字、布尔值、JSON等）
- **GET请求增强支持**：
  - 支持在请求体中放置JSON和XML数据（不再限制为仅查询参数）
  - 支持通过`--query`参数添加URL查询参数
  - 可以同时使用请求体数据和URL查询参数
- **鉴权机制支持**：
  - **Bearer Token认证**：通过 `--auth-bearer` 参数提供Token，自动添加 `Authorization: Bearer <token>` 头
  - **Basic Auth认证**：通过 `--auth-basic` 参数提供用户名密码（格式：`username:password`），自动进行Base64编码并添加 `Authorization: Basic <encoded>` 头
  - **API Key认证**：通过自定义header方式添加API Key，支持任意头名称
  - **组合使用**：支持多种鉴权方式和自定义头的组合使用
- **URL查询参数支持**：
  - **命令行参数**：通过 `--query` 或 `-q` 参数添加URL查询参数，格式为 `key=value`
  - **配置文件支持**：支持在配置文件的 `[request].query` 数组中定义查询参数
  - **参数优先级**：命令行参数优先级高于配置文件参数
  - **多参数支持**：可以使用多个 `--query` 参数添加多组查询参数
  - **适用范围**：所有HTTP请求方法（GET、POST等）都支持查询参数
- **自定义HTTP头支持**：
  - 通过 `--header` 参数添加自定义HTTP请求头
  - 支持添加多个header，每个header使用独立的 `--header` 参数
  - **严格格式要求**：`--header "HeaderName: HeaderValue"`
  - **格式验证**：
    - 必须包含冒号分隔符
    - 头名称不能为空
    - 格式错误时立即报错并停止执行
  - **错误处理**：
    - 缺少冒号：`自定义HTTP头格式错误: <input>，正确格式应为 'HeaderName: HeaderValue'`
    - 空头名称：`自定义HTTP头名称不能为空: <input>`
- 详细的执行结果显示，包括状态码、响应时间等
- 支持将测试结果保存为CSV文件
- 提供完整的统计信息（总数、成功数、失败数、成功率等）
- **调试模式支持**：
  - 使用 `--debug` 参数启用调试模式
  - **请求阶段**：格式化输出每个请求的URL、HTTP方法、超时时间、HTTP头和请求体内容
  - **响应阶段**：无论测试用例成功或失败，都会详细输出响应信息，包括：
    - 测试用例ID和基本信息（状态码、耗时、执行结果）
    - 错误信息（如果存在）
    - 完整的响应体内容（自动格式化JSON响应，便于阅读）
  - 便于问题排查、调试和结果分析

### 2.4 CSV文件格式约束

#### 2.4.1 生成阶段的格式约束

**XML格式生成：**
- `local-gen` 命令使用 `--xml` 参数时，生成单列CSV文件
- 列名固定为 "XML"
- 每行包含一个完整的XML字符串作为测试用例

**JSON格式生成：**
- `local-gen` 命令使用 `--json` 参数时，生成单列CSV文件
- 列名固定为 "JSON"
- 每行包含一个完整的JSON字符串作为测试用例

#### 2.4.2 使用阶段的格式识别

**自动格式检测：**
- `request` 命令能够从CSV文件第一行自动识别请求体格式类型
- **检测规则**：
  - 当第一行包含 "xml"（不区分大小写）时，自动识别为XML格式
  - 当第一行包含 "json"（不区分大小写）时，自动识别为JSON格式
  - 检测成功后无需手动指定 `--xml` 或 `--json` 参数
- **手动覆盖**：使用 `--xml` 或 `--json` 参数可以覆盖自动检测结果

**格式处理逻辑：**
- **单列XML格式**：当CSV文件只有一列且列名为 "XML" 时，直接使用每行的XML内容作为请求体
- **单列JSON格式**：当CSV文件只有一列且列名为 "JSON" 时，直接使用每行的JSON内容作为请求体
- **多列格式**：当CSV文件有多列时，将各列数据组合为JSON对象作为请求体

**请求方法约束：**
- **POST请求**：支持XML和JSON两种格式的CSV文件，数据放置在请求体中
- **GET请求**：支持XML和JSON两种格式的CSV文件，数据放置在请求体中（不再限制为仅JSON格式）
- **URL查询参数**：所有请求方法都支持通过`--query`参数添加URL查询参数

### 2.5 统一配置文件系统

#### 2.5.1 配置文件整合

为了简化配置管理和提高用户体验，系统实现了统一的配置文件管理系统：

**整合内容：**
- `[llm]`：LLM API配置
- `[request]`：请求相关配置
- `[testcase]`：用例设置（原common部分）
- `[constraints]`：约束配置
- `[builtin_data]`：内置数据

#### 2.5.2 完整配置文件结构

```toml
# API自动化测试命令行工具统一配置文件

[llm]
url = "https://api.llm.ai/v1"
api_key = "app-your-api-key-here"
user_prompt = "请生成边界情况和异常情况的测试用例"

[request]
url = "https://api.example.com/test"
method = "post"
file = "test_cases.csv"
save_path = "results.csv"
timeout = 30
concurrent = 3
auth_bearer = "your_bearer_token_here"
auth_basic = "username:password"
auth_api_key = "your_api_key_here"
headers = [
    "X-API-Version: v1",
    "X-Client-ID: atc-tool",
    "X-Request-Source: automated-test"
]

[testcase]
num = 10
output = "test_cases.csv"
type = "json"
# 随机化因子，控制数据变化程度（0.0-1.0，默认0.5）
variation_rate = 0.5
positive_example = '''
{
  "user": {
    "name": "张三",
    "age": 25,
    "phone": "13800138000"
  }
}
'''

[constraints.name]
type = "chinese_name"

[constraints.age]
type = "integer"
min = 1
max = 120

[builtin_data]
first_names = ["张", "王", "李"]
last_names = ["伟", "芳", "娜"]
email_domains = ["qq.com", "163.com"]
```

#### 2.5.3 公共参数统一

以下参数在所有相关命令中保持一致：

| 参数 | 短参数 | 说明 | 适用命令 | 配置文件位置 |
|------|--------|------|----------|-------------|
| `--config` | `-c` | 配置文件路径 | 所有命令 | - |
| `--num` | `-n` | 生成数量 | llm-gen, local-gen | `[testcase].num` |
| `--output` | `-o` | 输出文件路径 | llm-gen, local-gen | `[testcase].output` |
| `--ignore-tls` | - | 忽略TLS证书验证错误 | request, local-gen --exec, llm-gen --exec | `[request].ignore_tls_errors` |

**testcase配置参数详解：**

| 参数名 | 类型 | 默认值 | 说明 | 取值范围 |
|--------|------|--------|------|----------|
| `num` | 整数 | 10 | 生成测试用例数量 | 1-1000 |
| `output` | 字符串 | "test_cases.csv" | 输出文件路径 | 有效文件路径 |
| `type` | 字符串 | "json" | 正例报文类型 | "xml", "json" |
| `variation_rate` | 浮点数 | 0.5 | 随机化因子，控制数据变化程度 | 0.0-1.0 |
| `positive_example` | 字符串 | - | 正例报文内容 | 有效的XML或JSON |

**variation_rate参数说明：**
- `0.0`：不进行任何随机化，保持原值
- `0.1-0.3`：低度随机化，适用于需要保持数据相对稳定的场景
- `0.4-0.6`：中度随机化（默认0.5），平衡变化程度和数据合理性
- `0.7-1.0`：高度随机化，适用于边界测试和异常情况测试

#### 2.5.4 配置文件优先级

系统采用以下优先级规则：

1. **命令行显式参数**（最高优先级）
2. **配置文件参数**
3. **默认值**（最低优先级）

**智能参数合并：**
- 只有当参数为默认值时才从配置文件读取
- 避免配置文件意外覆盖用户显式设置的参数
- 保持命令行的灵活性

#### 2.5.5 多组HTTP头支持

配置文件支持TOML数组格式的多组HTTP头：

```toml
[request]
headers = [
    "X-API-Version: v1",
    "X-Client-ID: atc-tool",
    "Authorization: Bearer token123",
    "Content-Type: application/json"
]
```

## 3. 技术要求

### 3.1 开发环境

- 编程语言：Go 1.25
- 命令行框架：Cobra 1.10.1
- 配置文件解析：go-toml/v2（用于约束配置文件解析）

### 3.2 系统环境要求

**UTF-8编码支持**：
- 本工具在输出中使用emoji表情符号（✅、❌、🔍、📊、🎯、🧪、📚、🔧等）以提供更好的用户体验
- 要求终端环境必须支持UTF-8编码才能正确显示这些字符
- **Windows系统**：建议使用Windows Terminal、VSCode和IDEA内置的终端，使用Powershell或者cmd会导致emoji符号无法正常显示
- **macOS/Linux系统**：大多数现代终端默认支持UTF-8编码
- **SSH/远程连接**：确保SSH客户端和服务器都支持UTF-8编码
- 如果emoji字符显示异常，用户需要检查终端编码设置

### 3.3 性能要求

- 支持并发请求以提高测试效率
- 内存占用合理，避免大量数据导致内存溢出

### 3.4 可用性要求

- 命令行参数设计符合直觉，易于记忆
- 提供详细的帮助信息
- 错误信息清晰明确

### 3.5 技术限制

#### 3.5.1 XML编码支持限制

**重要说明：** Go标准库的XML处理包（`encoding/xml`）对XML文档编码有以下限制：

- **仅支持UTF-8编码**：Go标准库只能正确解析UTF-8编码的XML文档
- **不支持其他编码**：对于GBK、GB2312、ISO-8859-1等非UTF-8编码的XML文档，标准库无法直接处理
- **编码声明忽略**：即使XML文档声明了`<?xml version="1.0" encoding="GBK"?>`，Go标准库也会按UTF-8处理

**解决方案：**

本工具通过以下方式解决编码兼容性问题：

1. **自动编码检测**：使用`golang.org/x/text/encoding`包检测XML文档的实际编码
2. **编码转换**：将非UTF-8编码的XML文档自动转换为UTF-8编码后再进行解析
3. **支持的编码格式**：
   - UTF-8（原生支持）
   - GBK/GB2312（中文编码）
   - ISO-8859-1（西欧编码）
   - 其他常见编码格式

**使用建议：**

- **推荐使用UTF-8编码**：为获得最佳性能和兼容性，建议使用UTF-8编码的XML文档
- **非UTF-8编码处理**：工具会自动处理非UTF-8编码，但可能会有轻微的性能开销
- **编码声明一致性**：确保XML文档的编码声明与实际文件编码一致，避免解析错误

## 4. 交付物

- 源代码
- 使用说明文档
- 可执行文件（支持多平台）
  - Windows amd64
  - macOS ARM64
  - Linux ARM64
  - Linux amd64
