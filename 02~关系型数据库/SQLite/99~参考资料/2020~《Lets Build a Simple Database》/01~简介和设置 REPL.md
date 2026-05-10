# 简介和设置 REPL

作为一个 web 开发者，我每天都在工作中和关系型数据库打交道，但是他们对我来说是一个黑盒。这导致我出现了一些疑问：

- 数据保存的格式是什么？（在内存和硬盘中）
- 什么是候将它从内存移动到硬盘中？
- 为什么一个数据表只能有一个主键？
- 回滚操作是怎么完成的？
- 索引是如何格式化的？
- 何时及如何进行全表扫描？
- 准备好的语句以什么格式保存？

换句话说，一个数据库是如何**工作**的？

为了搞清楚这些事，我正在从头编写数据库。它是基于 sqlite 架构的，因为它的体积小，不像 MySQL 或是 PostgreSQL 那样，所以我更有希望理解它。整个数据库都存在一个.db 文件中！

# Sqlite

他们的网站上有许多 [内部 sqlite 文档](https://www.sqlite.org/arch.html)，另外，我还有一个 [SQLite 数据库系统：设计和实现](https://play.google.com/store/books/details?id=9Z6IQQnX1JEC) 的副本。

[![sqlite 架构（https://www.sqlite.org/zipvfs/doc/trunk/www/howitworks.wiki）](https://camo.githubusercontent.com/eb5784f48f09a00ed7834cc23ab3c92eb7137362ab9e099363972cbf5fc78848/68747470733a2f2f63737461636b2e6769746875622e696f2f64625f7475746f7269616c2f6173736574732f696d616765732f61726368312e676966)](https://camo.githubusercontent.com/eb5784f48f09a00ed7834cc23ab3c92eb7137362ab9e099363972cbf5fc78848/68747470733a2f2f63737461636b2e6769746875622e696f2f64625f7475746f7269616c2f6173736574732f696d616765732f61726368312e676966)

查询通过一系列组件来检索或是修改数据。**前端**包括：

- Tokenizer
- 解析器
- 代码生成器

输入到前端的是一个 SQL 查询。输出是 sqlite 虚拟机字节码（本质上是可以在数据库上运行的已编译程序）

*后端*包括：

- 虚拟机
- B 树
- Pager
- 系统接口

**虚拟机**将前端生成的字节码当成指令。然后，他可以对一个或多个表进行索引操作，每个表都存储在一种叫做 B 树的数据结构中。VM 本质上是关于字节码指令类型的大开关语句。

每个**B 树**都由许多个节点组成。每个节点的长度为一页。B 数可以通过向 Pager 发出命令来从硬盘检索页面或将其保存回硬盘。

**Pager** 接收命令以读取或写入数据页。它负责以适当的偏移量在数据库文件中进行读/写操作。它还会在内存中保留最近访问页面的缓存，并确定何时需要将这些页面写回到硬盘。

**系统接口**是不同的层，具体取决于为哪个操作系统的 sqlite 进行编译。在这个指南中，我不会去支持多种平台。

千里之行始于足下，让我们从一个简单的东西开始：**REPL**。

# 制作一个简单的 REPL

当你在命令行启动 Sqlite 的时候，Sqlite 会启动一个 read-execute-print loop：

```
~ sqlite3
SQLite version 3.16.0 2016-11-04 19:09:39
Enter ".help" for usage hints.
Connected to a transient in-memory database.
Use ".open FILENAME" to reopen on a persistent database.
sqlite> create table users (id int, username varchar(255), email varchar(255));
sqlite> .tables
users
sqlite> .exit
~
```

为此，我们的 `main` 函数应有一个无限循环，这个循环打印提示，获取输入，然后处理输入：

```c
int main(int argc, char* argv[]) {
  InputBuffer* input_buffer = new_input_buffer();
  while (true) {
    print_prompt();
    read_input(input_buffer);

    if (strcmp(input_buffer->buffer, ".exit") == 0) {
      close_input_buffer(input_buffer);
      exit(EXIT_SUCCESS);
    } else {
      printf("Unrecognized command '%s'.\n", input_buffer->buffer);
    }
  }
}
```

我们将 `InputBuffer` 定义为一个围绕状态存储的小包装，以便和 [getline()](http://man7.org/linux/man-pages/man3/getline.3.html) 函数进行交互（稍后详细介绍）

```c
typedef struct {
  char* buffer;
  size_t buffer_length;
  ssize_t input_length;
} InputBuffer;

InputBuffer* new_input_buffer() {
  InputBuffer* input_buffer = (InputBuffer*)malloc(sizeof(InputBuffer));
  input_buffer->buffer = NULL;
  input_buffer->buffer_length = 0;
  input_buffer->input_length = 0;

  return input_buffer;
}
```

接着，`print_prompt()` 打印一个提示给用户。我们在读每一行输入之前都要进行这个操作。

```c
void print_prompt() { printf("db > "); }
```

如果要读取一行输入，那就使用 [getline()](http://man7.org/linux/man-pages/man3/getline.3.html)：

```
ssize_t getline(char **lineptr, size_t *n, FILE *stream);
```

`lineptr`：一个指向字符串的指针，我们用来指向包含读取行的缓冲区。如果将它设置为 NULL，那它就将由 getline 分配，因此即使命令失败，也应由用户释放。

`n`：一个指向用于保存为缓冲区分配的大小的变量的指针。

`stream`：输入来源。我们使用标准输入。

`返回值`：读取的字节数，可能小于缓冲区的大小。

我们让 `getline` 将读取的行存储在 `input_buffer->buffer` 中，将分配的缓冲区大小存储在 `input_buffer->buffer_length` 中。我们将返回值存储在 `input_buffer->input_length` 中。

`buffer` 开始为空，所以 `getline` 分配足够的内存来容纳输入行，并让缓冲区指向输入行。

```c
void read_input(InputBuffer* input_buffer) {
  ssize_t bytes_read =
      getline(&(input_buffer->buffer), &(input_buffer->buffer_length), stdin);

  if (bytes_read <= 0) {
    printf("Error reading input\n");
    exit(EXIT_FAILURE);
  }

  // Ignore trailing newline
  input_buffer->input_length = bytes_read - 1;
  input_buffer->buffer[bytes_read - 1] = 0;
}
```

现在定义一个释放为 `InputBuffer *` 实例分配的内存和响应结构的 `buffer` 元素（`getline` 在 `read_input` 中为 `input_buffer->buffer` 分配的内存）

```c
void close_input_buffer(InputBuffer* input_buffer) {
    free(input_buffer->buffer);
    free(input_buffer);
}
```

最后，我们解析并执行命令。现在只有一个可识别的命令：.exit，它将终止程序。否则，我们将显示错误消息并继续循环。

```c
if (strcmp(input_buffer->buffer, ".exit") == 0) {
  close_input_buffer(input_buffer);
  exit(EXIT_SUCCESS);
} else {
  printf("Unrecognized command '%s'.\n", input_buffer->buffer);
}
```

来试试！

```
~ ./db
db > .tables
Unrecognized command '.tables'.
db > .exit
~
```

好了，我们有一个有效的 REPL。在下一部分中，我们将开始开发命令语言。同时，这是此部分中的整个程序：

```c
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
  char* buffer;
  size_t buffer_length;
  ssize_t input_length;
} InputBuffer;

InputBuffer* new_input_buffer() {
  InputBuffer* input_buffer = malloc(sizeof(InputBuffer));
  input_buffer->buffer = NULL;
  input_buffer->buffer_length = 0;
  input_buffer->input_length = 0;

  return input_buffer;
}

void print_prompt() { printf("db > "); }

void read_input(InputBuffer* input_buffer) {
  ssize_t bytes_read =
      getline(&(input_buffer->buffer), &(input_buffer->buffer_length), stdin);

  if (bytes_read <= 0) {
    printf("Error reading input\n");
    exit(EXIT_FAILURE);
  }

  // Ignore trailing newline
  input_buffer->input_length = bytes_read - 1;
  input_buffer->buffer[bytes_read - 1] = 0;
}

void close_input_buffer(InputBuffer* input_buffer) {
    free(input_buffer->buffer);
    free(input_buffer);
}

int main(int argc, char* argv[]) {
  InputBuffer* input_buffer = new_input_buffer();
  while (true) {
    print_prompt();
    read_input(input_buffer);

    if (strcmp(input_buffer->buffer, ".exit") == 0) {
      close_input_buffer(input_buffer);
      exit(EXIT_SUCCESS);
    } else {
      printf("Unrecognized command '%s'.\n", input_buffer->buffer);
    }
  }
}
```
