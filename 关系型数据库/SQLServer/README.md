SQL Server 2016 有望提供 JSON 操作原生支持。这一支持的首次迭代将作为 SQL Server 2016 CTP 2 的一部分发布。CTP 又名社区技术预览版，等同于微软的公开 Alpha 测试，期间，开发者可以提出技术上的修改建议。

**JSON 导出**

“格式化和导出”JSON 是 CTP 2 的主要功能。这可以通过在 SELECT 语句的末尾增加 FOR JSON 子句实现。该功能基于 FOR XML 子句，并且与它类似，允许对结果 JSON 字符串进行自动和半手工格式化。

微软希望这一语法可以提供与 PostgreSQL 的 row_to_json 和 json_object 函数相同的功能。

**JSON 转换**

大部分功能将在 CTP 3 中发布。这些功能中的第一项是 FROM OPENJSON 子句，这是一个表值函数(TFV)，它接受一个 JSON 字符串作为输入。它还需要一个数组或对象在被解析的 JSON 数据中的路径。

默认情况下，OPENJSON 返回一个键值对集合，但开发者可以使用 WITH 子句提供一个模式。由于 JSON 不支持日期或整数(它们分别表示为字符串和 double 类型)，所以 WITH 子句可以减少稍后需要的类型转换数量。

下面是[Jovan Popovic 博文中关于 JSON 支持](http://blogs.msdn.com/b/jocapc/archive/2015/05/16/json-support-in-sql-server-2016.aspx)的例子。

```
DECLARE @JSalestOrderDetails nVarCar(2000) = N '{"OrdersArray": [
{"Number":1, "Date": "8/10/2012", "Customer": "Adventure works", "Quantity": 1200},
{"Number":4, "Date": "5/11/2012", "Customer": "Adventure works", "Quantity": 100},
{"Number":6, "Date": "1/3/2012", "Customer": "Adventure works", "Quantity": 250},
{"Number":8, "Date": "12/7/2012", "Customer": "Adventure works", "Quantity": 2200}
]}';

SELECT Number, Customer, Date, Quantity
FROM OPENJSON (@JSalestOrderDetails, '$.OrdersArray')
WITH (
    Number varchar(200),
    Date datetime,
    Customer varchar(200),
    Quantity int
) AS OrdersArray
```

微软宣称，在 PostgrSQL 中实现同样的功能需要综合使用 json_each、json_object_keys、json_populate_record 和 json_populate_recordset 函数。

**JSON 存储**

正如所见，JSON 数据存储在 NVARCHAR 变量中。在表的列中存储 JSON 数据也是这样。关于为什么这么做，微软有如下理由：

> - 迁移——我们发现，人们早已将 JSON 存储为文本，因此，如果我们引入一种单独的 JSON 类型，那么为了使用我们的新特性，他们将需要修改数据库模式，并重新加载数据。而采用现在这种实现方式，开发者不需要对数据库做任何修改就可以使用 JSON 功能。
> - 跨功能的兼容性——所有 SQL Server 组件均支持 NVARCHAR，因此，JSON 也将在所有的组件中得到支持。开发者可以将 JSON 存储在 Hekaton、时态表或列存储表中，运用包括行级安全在内的标准安全策略，使用标准 B 树索引和 FTS 索引，使用 JSON 作为参数或返回过程值，等等。开发者不需要考虑功能 X 是否支持 JSON——如果功能 X 支持 NVARCHAR，那么它也支持 JSON。此外，该特性有一些约束——Hekaton 及列存储不支持 LOB 值，所以开发者只能存储小 JSON 文档。不过，一旦我们在 Hekaton 及列存储中增加了 LOB 支持，那么开发者就可以在任何地方存储大 JSON 文档了。
> - 客户端支持——目前，我们没有为客户端应用程序提供标准 JSON 对象类型(类似 XmlDom 对象的东西)。自然地，Web 和移动应用程序以及 JavaScript 客户端将使用 JSON 文本，并使用本地解析器解析它。在 JavaScript 中，可以使用 object 类型表示 JSON。我们不太可能实现一些仅在少数 RDBMS 中存在的 JSON 类型代理。在 C#.Net 中，许多开发者使用内置了 JObject 或 JArray 类型的 JSON.Net 解析器；不过，那不是一种标准，也不太可能成为 ADO.NET 的一部分。即便如此，我们认为，C#应用可以接受来自数据库层的纯字符串，并使用最喜欢的解析器解析它。我们所谈论的内容不只是跟应用程序有关。如果开发者试图在 SSIS/SSRS、Tableau 和 Informatica ETL 中使用 JSON 列，它们会将其视为文本。我们认为，即使我们增加了 JSON 类型，在 SQL Server 之外，它仍将被表示成字符串，并根据需要使用某个自定义的解析器解析它。因此，我们并没有找到任何重大的理由将其实现为一种原生 JSON 类型。

在包含 JSON 的 NVARCHAR 列上使用新的 ISJSON 函数作为检查约束是个不错的主意。如果不这样做，那么有缺陷的客户端应用程序就可能插入不可解析的字符串，使开发者面临数据污染的风险。

**JSON 查询**

如果直接对 JSON 进行标量查询，可以使用 JSON_VALUE 函数。该函数使用一种类似 JavaScript 的符号定位 JSON 对象中的值。它使用\$符号表示 object 的根，点号表示属性，方括号表示数组索引。它与 PostgreSQL 中的 json_extract_path_text 函数等效。

**JSON 索引**

JSON 数据可以直接索引，但开发者可以毫不费力地在标量数据上实现同样的效果。只需要使用 JSON_VALUE 函数创建一个计算列，然后在这个列上创建索引。
