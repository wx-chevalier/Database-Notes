> [原文地址](https://xata.io/blog/postgres-full-text-search-engine)

# Create an advanced search engine with PostgreSQL

This is part 1 of a blog mini-series, in which we explore the full-text search functionality in PostgreSQL and investigate how much of the typical search engine functionality we can replicate. In part 2, we'll do a comparison between PostgreSQL's full-text search and Elasticsearch.
这是博客迷你系列的第 1 部分，其中我们探讨了 PostgreSQL 中的全文搜索功能，并调查了我们可以复制多少典型的搜索引擎功能。在第 2 部分中，我们将比较 PostgreSQL 的全文搜索和 Elasticsearch。

If you want to follow along and try out the sample queries (which we recommend; it's more fun that way), the code samples are executed against the [Wikipedia Movie Plots](https://www.kaggle.com/datasets/jrobischon/wikipedia-movie-plots) data set from Kaggle. To import it, download the CSV file, then create this table:
如果您想继续尝试示例查询（我们推荐这样做;这样更有趣），代码示例将针对 Kaggle 中的维基百科电影情节数据集执行。要导入它，请下载 CSV 文件，然后创建以下表：

```sql
CREATE TABLE movies(
	ReleaseYear int,
	Title text,
	Origin text,
	Director text,
	Casting text,
	Genre text,
	WikiPage text,
	Plot text);
```

And import the CSV file like this:
并像这样导入 CSV 文件：

```
\COPY movies(ReleaseYear, Title, Origin, Director, Casting, Genre, WikiPage, Plot)
	FROM 'wiki_movie_plots_deduped.csv' DELIMITER ',' CSV HEADER;
```

The dataset contains 34,000 movie titles and is about 81 MB in CSV format.
该数据集包含 34，000 个电影标题，CSV 格式约为 81 MB。

[#PostgreSQL full-text search primitives
PostgreSQL 全文搜索原语](https://xata.io/blog/postgres-full-text-search-engine#postgresql-full-text-search-primitives)

The Postgres approach to full-text search offers building blocks that you can combine to create your own search engine. This is quite flexible but it also means it generally feels lower-level compared to search engines like Elasticsearch, Typesense, or Mellisearch, for which full-text search is the primary use case.
Postgres 全文搜索方法提供了构建块，您可以组合这些构建块来创建自己的搜索引擎。这是非常灵活的，但这也意味着与 Elasticsearch，Typesense 或 Mellisearch 等搜索引擎相比，它通常感觉较低级别，其中全文搜索是主要用例。

The main building blocks, which we'll cover via examples, are:
我们将通过示例介绍的主要构建块是：

- The `tsvector` and `tsquery` data types
  和 `tsvector` `tsquery` 数据类型
- The match operator `@@` to check if a `tsquery` matches a `tsvector`
  用于检查 a `tsquery` 是否匹配的 `tsvector` 匹配运算符 `@@`
- Functions to rank each match (`ts_rank`, `ts_rank_cd`)
  对每个匹配项进行排名的函数 （ `ts_rank` ， `ts_rank_cd` ）
- The GIN index type, an inverted index to efficiently query `tsvector`
  GIN 索引类型，用于高效查询 `tsvector` 的倒排索引

We'll start by looking at these building blocks and then we'll get into more advanced topics, covering relevancy boosters, typo-tolerance, and faceted search.
我们将从查看这些构建块开始，然后进入更高级的主题，涵盖相关性助推器、拼写错误容错和分面搜索。

[#tsvector 向量](https://xata.io/blog/postgres-full-text-search-engine#tsvector)

The `tsvector` data type stores a sorted list of _lexemes_. A _lexeme_ is a string, just like a token, but it has been _normalized_ so that different forms of the same word are made. For example, normalization almost always includes folding upper-case letters to lower-case, and often involves removal of suffixes (such as `s` or `ing` in English). Here is an example, using the `to_tsvector` function to parse an English phrase into a `tsvector`.
数据类型 `tsvector` 存储词素的排序列表。词素是一个字符串，就像一个标记一样，但它已被规范化，以便制作同一单词的不同形式。例如，规范化几乎总是包括将大写字母折叠为小写字母，并且通常涉及删除后缀（例如 `s` 或 `ing` 英语）。下面是一个示例，使用该 `to_tsvector` 函数将英语短语解析为 `tsvector` .

```
SELECT * FROM unnest(to_tsvector('english',
	'I''m going to make him an offer he can''t refuse. Refusing is not an option.'));

 lexeme | positions | weights
--------+-----------+---------
 go     | {3}       | {D}
 m      | {2}       | {D}
 make   | {5}       | {D}
 offer  | {8}       | {D}
 option | {17}      | {D}
 refus  | {12,13}   | {D,D}
(6 rows)
```

As you can see, stop words like “I”, “to” or “an” are removed, because they are too common to be useful for search. The words are normalized and reduced to their root (e.g. “refuse” and “Refusing” are both transformed into “refus”). The punctuation signs are ignored. For each word, the **positions** in the original phrase are recorded (e.g. “refus” is the 12th and the 13th word in the text) and the **weights** (which are useful for ranking and we'll discuss later).
如您所见，删除了诸如“I”，“to”或“an”之类的停用词，因为它们太常见而无法用于搜索。这些词被规范化并简化为词根（例如，“拒绝”和“拒绝”都转换为“拒绝”）。标点符号将被忽略。对于每个单词，记录原始短语中的位置（例如，“refus”是文本中的第 12 个和第 13 个单词）和权重（这对排名很有用，我们将在后面讨论）。

In the example above, the transformation rules from words to _lexemes_ are based on the `english` search configuration. Running the same query with the `simple` search configuration results in a `tsvector` that includes all the words as they were found in the text:
在上面的示例中，从单词到词素的转换规则基于 `english` 搜索配置。使用 `simple` 搜索配置运行相同的查询会产生 ， `tsvector` 其中包含在文本中找到的所有单词：

```
SELECT * FROM unnest(to_tsvector('simple',
	'I''m going to make him an offer he can''t refuse. Refusing is not an option.'));

  lexeme  | positions | weights
----------+-----------+---------
 an       | {7,16}    | {D,D}
 can      | {10}      | {D}
 going    | {3}       | {D}
 he       | {9}       | {D}
 him      | {6}       | {D}
 i        | {1}       | {D}
 is       | {14}      | {D}
 m        | {2}       | {D}
 make     | {5}       | {D}
 not      | {15}      | {D}
 offer    | {8}       | {D}
 option   | {17}      | {D}
 refuse   | {12}      | {D}
 refusing | {13}      | {D}
 t        | {11}      | {D}
 to       | {4}       | {D}
(16 rows)
```

As you can see, “refuse” and “refusing” now result in different lexemes. The `simple` configuration is particularly useful when you have columns that contain labels or tags.
如您所见，“拒绝”和“拒绝”现在会导致不同的词素。当您的列包含标签或标记时，此 `simple` 配置特别有用。

PostgreSQL has built-in configurations for a pretty good set of languages. You can see the list by running:
PostgreSQL 为一组相当不错的语言提供了内置配置。您可以通过运行以下命令查看列表：

```
SELECT cfgname FROM pg_ts_config;
```

Notably, however, there is no configuration for CJK (Chinese-Japanese-Korean), which is worth keeping in mind if you need to create a search query in those languages. While the `simple` configuration should work in practice quite well for unsupported languages, I'm not sure if that is enough for CJK.
但是，值得注意的是，CJK（中文-日语-韩语）没有配置，如果您需要使用这些语言创建搜索查询，则值得记住这一点。虽然该 `simple` 配置在实践中应该适用于不受支持的语言，但我不确定这对于 CJK 来说是否足够。

[#tsquery 查询](https://xata.io/blog/postgres-full-text-search-engine#tsquery)

The `tsquery` data type is used to represent a normalized query. A `tsquery` contains search terms, which must be already-normalized lexemes, and may combine multiple terms using AND, OR, NOT, and FOLLOWED BY operators. There are functions like `to_tsquery`, `plainto_tsquery`, and `websearch_to_tsquery` that are helpful in converting user-written text into a proper `tsquery`, primarily by normalizing words appearing in the text.
数据类型 `tsquery` 用于表示规范化查询。A `tsquery` 包含搜索词，这些词必须是已经规范化的词素，并且可以使用 AND、OR、NOT 和后跟运算符组合多个词。有一些函数，如 ， 和 `websearch_to_tsquery` 有助于将用户编写的文本转换为适当的 `tsquery` 文本 `to_tsquery` ， `plainto_tsquery` 主要是通过规范化文本中出现的单词。

To get a feeling of `tsquery`, let's see a few examples using `websearch_to_tsquery`:
为了获得一种 `tsquery` 感觉，让我们看几个使用 `websearch_to_tsquery` 的例子：

```
SELECT websearch_to_tsquery('english', 'the dark vader');
 websearch_to_tsquery
----------------------
'dark' & 'vader'
```

That is a logical AND, meaning that the document needs to contain both “quick” and “dog” in order to match. You can do logical OR as well:
这是一个合乎逻辑的 AND，这意味着文档需要同时包含“快速”和“狗”才能匹配。您也可以执行逻辑或：

```
SELECT websearch_to_tsquery('english', 'quick OR dog');
 websearch_to_tsquery
----------------------
 'dark' | 'vader'
```

And you can exclude words:
您可以排除字词：

```
SELECT websearch_to_tsquery('english', 'dark vader -wars');   websearch_to_tsquery--------------------------- 'dark' & 'vader' & !'war'
```

Also, you can represent phrase searches:
此外，您还可以表示短语搜索：

```
SELECT websearch_to_tsquery('english', 'dark vader -wars');
   websearch_to_tsquery
---------------------------
 'dark' & 'vader' & !'war'
```

This means: “dark”, followed by “vader”, followed by “son”.
这意味着：“黑暗”，然后是“维达”，然后是“儿子”。

Note, however, that the “the” word is ignored, because it's a stop word as per the `english` search configuration. This can be an issue on phrases like this:
但请注意，“the”字将被忽略，因为根据 `english` 搜索配置，它是停用词。这可能是这样的短语的问题：

```
SELECT websearch_to_tsquery('english', '"do or do not, there is no try"');
 websearch_to_tsquery
----------------------
 'tri'
(1 row)
```

Oops, almost the entire phrase was lost. Using the `simple` config gives the expected result:
哎呀，几乎整个短语都丢失了。使用 `simple` 配置给出预期的结果：

```
SELECT websearch_to_tsquery('simple', '"do or do not, there is no try"');
                           websearch_to_tsquery
--------------------------------------------------------------------------
 'do' <-> 'or' <-> 'do' <-> 'not' <-> 'there' <-> 'is' <-> 'no' <-> 'try'
```

You can check whether a `tsquery` matches a `tsvector` by using the match operator `@@`.
您可以使用匹配运算符 `@@` 检查 a `tsquery` `tsvector` 是否匹配。

```
SELECT websearch_to_tsquery('english', 'dark vader') @@
	to_tsvector('english',
		'Dark Vader is my father.');

?column?
----------
 t
```

While the following example doesn't match:
虽然以下示例不匹配：

```
SELECT websearch_to_tsquery('english', 'dark vader -father') @@
	to_tsvector('english',
		'Dark Vader is my father.');

?column?
----------
 f
```

[#GIN](https://xata.io/blog/postgres-full-text-search-engine#gin)

Now that we've seen `tsvector` and `tsquery` at work, let's look at another key building block: the GIN index type is what makes it fast. GIN stands for _Generalized Inverted Index._ GIN is designed for handling cases where the items to be indexed are composite values, and the queries to be handled by the index need to search for element values that appear within the composite items. This means that GIN can be used for more than just text search, notably for JSON querying.
现在我们已经看到 `tsvector` 并 `tsquery` 正在工作，让我们看看另一个关键的构建块：GIN 索引类型是使其快速的原因。GIN 代表 广义倒排指数。GIN 设计用于处理要索引的项目是复合值的情况，索引要处理的查询需要搜索复合项中出现的元素值。这意味着 GIN 不仅可以用于文本搜索，尤其是 JSON 查询。

You can create a GIN index on a set of columns, or you can first create a column of type `tsvector`, to include all the searchable columns. Something like this:
可以在一组列上创建 GIN 索引，也可以先创建 类型的 `tsvector` 列，以包含所有可搜索的列。像这样：

```
ALTER TABLE movies ADD search tsvector GENERATED ALWAYS AS
	(to_tsvector('english', Title) || ' ' ||
   to_tsvector('english', Plot) || ' ' ||
   to_tsvector('simple', Director) || ' ' ||
	 to_tsvector('simple', Genre) || ' ' ||
   to_tsvector('simple', Origin) || ' ' ||
   to_tsvector('simple', Casting)
) STORED;
```

And then create the actual index:
然后创建实际索引：

```
CREATE INDEX idx_search ON movies USING GIN(search);
```

You can now perform a simple test search like this:
您现在可以执行简单的测试搜索，如下所示：

```
SELECT title FROM movies WHERE search @@ websearch_to_tsquery('english','dark vader');                         title-------------------------------------------------- Star Wars Episode IV: A New Hope (aka Star Wars) Return of the Jedi Star Wars: Episode III – Revenge of the Sith(3 rows)
```

To see the effects of the index, you can compare the timings of the above query with and without the index. The GIN index takes it from over 200 ms to about 4 ms on my computer.
若要查看索引的效果，可以比较使用和不使用索引的上述查询的计时。GIN 索引在我的计算机上从 200 多毫秒增加到大约 4 毫秒。

[#ts_rank ts_rank](https://xata.io/blog/postgres-full-text-search-engine#tsrank)

So far, we've seen how `ts_vector` and `ts_query` can match search queries. However, for a good search experience, it is important to show the best results first - meaning that the results need to be sorted by _relevancy_.
到目前为止，我们已经了解了如何 `ts_vector` 以及如何 `ts_query` 匹配搜索查询。但是，为了获得良好的搜索体验，首先显示最佳结果非常重要 - 这意味着需要按相关性对结果进行排序。

Taking it directly from the [docs](https://www.postgresql.org/docs/current/textsearch-controls.html#TEXTSEARCH-RANKING):
直接从文档中获取：

> PostgreSQL provides two predefined ranking functions, which take into account lexical, proximity, and structural information; that is, they consider how often the query terms appear in the document, how close together the terms are in the document, and how important is the part of the document where they occur. However, the concept of relevancy is vague and very application-specific. Different applications might require additional information for ranking, e.g., document modification time. The built-in ranking functions are only examples. You can write your own ranking functions and/or combine their results with additional factors to fit your specific needs.
> PostgreSQL 提供了两个预定义的排名函数，它们考虑了词汇、邻近和结构信息;也就是说，它们考虑查询词在文档中出现的频率、这些词在文档中的接近程度以及它们出现的文档部分的重要性。但是，相关性的概念是模糊的，并且非常特定于应用程序。不同的应用程序可能需要额外的排名信息，例如文档修改时间。内置排名函数只是示例。您可以编写自己的排名函数和/或将其结果与其他因素相结合，以满足您的特定需求。

The two ranking functions mentioned are `ts_rank` and `ts_rank_cd`. The difference between them is that while they both take into account the frequency of the term, `ts_rank_cd` also takes into account the proximity of matching lexemes to each other.
提到的两个排名函数是 `ts_rank` 和 `ts_rank_cd` 。它们之间的区别在于，虽然它们都考虑了项的频率， `ts_rank_cd` 但也考虑了匹配词素彼此之间的接近程度。

To use them in a query, you can do something like this:
若要在查询中使用它们，可以执行以下操作：

```
SELECT title,
       ts_rank(search, websearch_to_tsquery('english', 'dark vader')) rank
  FROM movies
  WHERE search @@ websearch_to_tsquery('english','dark vader')
  ORDER BY rank DESC
  LIMIT 10;

 title                                            |    rank
--------------------------------------------------+------------
 Return of the Jedi                               | 0.21563873
 Star Wars: Episode III – Revenge of the Sith     | 0.12592985
 Star Wars Episode IV: A New Hope (aka Star Wars) | 0.05174401
```

One thing to note about `ts_rank` is that it needs to access the `search` column for each result. This means that if the `WHERE` condition matches a lot of rows, PostgreSQL needs to visit them all in order to do the ranking, and that can be slow. To exemplify, the above query returns in 5-7 ms on my computer. If I modify the query to do search for `dark OR vader`, it returns in about 80 ms, because there are now over 1000 matching result that need ranking and sorting.
需要注意的一件事 `ts_rank` 是，它需要访问每个结果的 `search` 列。这意味着如果 `WHERE` 条件匹配很多行，PostgreSQL 需要访问所有行才能进行排名，这可能会很慢。例如，上述查询在我的计算机上在 5-7 毫秒内返回。如果我修改查询以执行搜索，它会在大约 80 毫秒内返回 `dark OR vader` ，因为现在有超过 1000 个匹配结果需要排名和排序。

# Relevancy tuning 相关性调整

While relevancy based on word frequency is a good default for the search sorting, quite often the data contains important indicators that are more relevant than simply the frequency.
虽然基于词频的相关性是搜索排序的良好默认值，但数据通常包含比频率更相关的重要指标。

Here are some examples for a movies dataset:
下面是电影数据集的一些示例：

- Matches in the title should be given higher importance than matches in the description or plot.
  标题中的匹配项应比描述或情节中的匹配项具有更高的重要性。
- More popular movies can be promoted based on ratings and/or the number of votes they receive.
  更受欢迎的电影可以根据评分和/或获得的票数进行推广。
- Certain categories can be boosted more, considering user preferences. For instance, if a particular user enjoys comedies, those movies can be given a higher priority.
  考虑到用户偏好，某些类别可以进一步提升。例如，如果特定用户喜欢喜剧，则可以赋予这些电影更高的优先级。
- When ranking search results, newer titles can be considered more relevant than very old titles.
  在对搜索结果进行排名时，可以认为较新的标题比非常旧的标题更相关。

This is why dedicated search engines typically offer ways to use different columns or fields to influence the ranking. Here are example tuning guides from [Elastic](https://www.elastic.co/guide/en/app-search/current/relevance-tuning-guide.html), [Typesense](https://typesense.org/docs/guide/ranking-and-relevance.html), and [Meilisearch](https://www.meilisearch.com/docs/learn/core_concepts/relevancy).
这就是为什么专用搜索引擎通常会提供使用不同列或字段来影响排名的方法。以下是来自 Elastic、Typesense 和 Meilisearch 的示例调优指南。

If you want a visual demo of the impact of relevancy tuning, here is a quick 4 minutes video about it:
如果您想直观地演示相关性调优的影响，这里有一个 4 分钟的快速视频：

<iframe src="https://www.youtube.com/embed/GfgdQs4WuXM" title="YouTube video player" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" style="border-width: 0px; border-style: solid; box-sizing: border-box; overflow-wrap: break-word; display: flex; border-color: var(--chakra-colors-chakra-border-color); overflow: hidden; position: absolute; inset: 0px; -webkit-box-pack: center; justify-content: center; -webkit-box-align: center; align-items: center; width: 800px; height: 450px;"></iframe>

[#Numeric, date, and exact value boosters
数字、日期和精确值助推器](https://xata.io/blog/postgres-full-text-search-engine#numeric-date-and-exact-value-boosters)

While Postgres doesn't have direct support for boosting based on other columns, the rank is ultimately just a sort expression, so you can add your own signals to it.
虽然 Postgres 没有直接支持基于其他列的提升，但排名最终只是一个排序表达式，因此您可以向其添加自己的信号。

For example, if you want to add a boost for the number of votes, you can do something like this:
例如，如果要增加票数，可以执行以下操作：

```
SELECT title,
  ts_rank(search, websearch_to_tsquery('english', 'jedi'))
    -- numeric booster example
    + log(NumberOfVotes)*0.01
 FROM movies
 WHERE search @@ websearch_to_tsquery('english','jedi')
 ORDER BY rank DESC LIMIT 10;
```

The logarithm is there to smoothen the impact, and the 0.01 factor brings the booster to a comparable scale with the ranking score.
对数是为了平滑影响，0.01 因子使助推器与排名分数相当。

You can also design more complex boosters, for example, boost by the rating, but only if the ranking has a certain number of votes. To do this, you can create a function like this:
您还可以设计更复杂的助推器，例如，通过评级提升，但前提是排名有一定数量的票数。为此，您可以创建如下函数：

```sql
create function numericBooster(rating numeric, votes numeric, voteThreshold numeric)
	returns numeric as $$
		select case when votes < voteThreshold then 0 else rating end;
$$ language sql;
```

And use it like this:
并像这样使用它：

```sql
SELECT title,
  ts_rank(search, websearch_to_tsquery('english', 'jedi'))
    -- numeric booster example
    + numericBooster(Rating, NumberOfVotes, 100)*0.005
 FROM movies
 WHERE search @@ websearch_to_tsquery('english','jedi')
 ORDER BY rank DESC LIMIT 10;
```

Let's take another example. Say we want to boost the ranking of comedies. You can create a `valueBooster` function that looks like this:
让我们再举一个例子。假设我们想提高喜剧的排名。您可以创建一个如下所示的 `valueBooster` 函数：

```
create function valueBooster (col text, val text, factor integer)
	returns integer as $$
		select case when col = val then factor else 0 end;
$$ language sql;
```

The function returns a factor if the column matches a particular value and 0 instead. Use it in a query like this:
如果列与特定值匹配，则该函数将返回一个因子，改为 0。在如下所示的查询中使用它：

```sql
SELECT title, genre,
   ts_rank(search, websearch_to_tsquery('english', 'jedi'))
   -- value booster example
   + valueBooster(Genre, 'comedy', 0.05) rank
FROM movies
   WHERE search @@ websearch_to_tsquery('english','jedi')                                                                                                 ORDER BY rank DESC LIMIT 10;
                      title                       |               genre                |        rank
--------------------------------------------------+------------------------------------+---------------------
 The Men Who Stare at Goats                       | comedy                             |  0.1107927106320858
 Clerks                                           | comedy                             |  0.1107927106320858
 Star Wars: The Clone Wars                        | animation                          | 0.09513916820287704
 Star Wars: Episode I – The Phantom Menace 3D     | sci-fi                             | 0.09471701085567474
 Star Wars: Episode I – The Phantom Menace        | space opera                        | 0.09471701085567474
 Star Wars: Episode II – Attack of the Clones     | science fiction                    | 0.09285612404346466
 Star Wars: Episode III – Revenge of the Sith     | science fiction, action            | 0.09285612404346466
 Star Wars: The Last Jedi                         | action, adventure, fantasy, sci-fi |  0.0889768898487091
 Return of the Jedi                               | science fiction                    | 0.07599088549613953
 Star Wars Episode IV: A New Hope (aka Star Wars) | science fiction                    | 0.07599088549613953
(10 rows)
```

[#Column weights 列重量](https://xata.io/blog/postgres-full-text-search-engine#column-weights)

Remember when we talked about the `tsvector` lexemes and that they can have weights attached? Postgres supports 4 weights, named A, B, C, and D. A is the biggest weight while D is the lowest and the default. You can control the weights via the `setweight` function which you would typically call when building the `tsvector` column:
还记得我们谈论词 `tsvector` 素并且它们可以附加权重吗？Postgres 支持 4 个权重，分别命名为 A、B、C 和 D。A 是最大权重，而 D 是最低和默认值。您可以通过构建 `tsvector` 列时通常会调用的 `setweight` 函数来控制权重：

```
ALTER TABLE movies ADD search tsvector GENERATED ALWAYS AS
	(to_tsvector('english', Title) || ' ' ||
   to_tsvector('english', Plot) || ' ' ||
   to_tsvector('simple', Director) || ' ' ||
	 to_tsvector('simple', Genre) || ' ' ||
   to_tsvector('simple', Origin) || ' ' ||
   to_tsvector('simple', Casting)
) STORED;
```

Let's see the effects of this. Without `setweight`, a search for `dark vader OR jedi` returns:
让我们看看这样做的效果。如果没有 ，则 `setweight` 搜索 `dark vader OR jedi` 返回：

```sql
SELECT title, ts_rank(search, websearch_to_tsquery('english', 'jedi')) rank
   FROM movies
   WHERE search @@ websearch_to_tsquery('english','jedi')
   ORDER BY rank DESC;
                      title                       |    rank
--------------------------------------------------+-------------
 Star Wars: The Clone Wars                        |  0.09513917
 Star Wars: Episode I – The Phantom Menace        |  0.09471701
 Star Wars: Episode I – The Phantom Menace 3D     |  0.09471701
 Star Wars: Episode III – Revenge of the Sith     | 0.092856124
 Star Wars: Episode II – Attack of the Clones     | 0.092856124
 Star Wars: The Last Jedi                         |  0.08897689
 Return of the Jedi                               | 0.075990885
 Star Wars Episode IV: A New Hope (aka Star Wars) | 0.075990885
 Clerks                                           |  0.06079271
 The Empire Strikes Back                          |  0.06079271
 The Men Who Stare at Goats                       |  0.06079271
 How to Deal                                      |  0.06079271
(12 rows)
```

And with the `setweight` on the title column:
并在标题列 `setweight` 上：

```sql
SELECT title, ts_rank(search, websearch_to_tsquery('english', 'jedi')) rank
   FROM movies
   WHERE search @@ websearch_to_tsquery('english','jedi')
   ORDER BY rank DESC;
                      title                       |    rank
--------------------------------------------------+-------------
 Star Wars: The Last Jedi                         |   0.6361112
 Return of the Jedi                               |   0.6231253
 Star Wars: The Clone Wars                        |  0.09513917
 Star Wars: Episode I – The Phantom Menace        |  0.09471701
 Star Wars: Episode I – The Phantom Menace 3D     |  0.09471701
 Star Wars: Episode III – Revenge of the Sith     | 0.092856124
 Star Wars: Episode II – Attack of the Clones     | 0.092856124
 Star Wars Episode IV: A New Hope (aka Star Wars) | 0.075990885
 The Empire Strikes Back                          |  0.06079271
 Clerks                                           |  0.06079271
 The Men Who Stare at Goats                       |  0.06079271
 How to Deal                                      |  0.06079271
(12 rows)
```

Note how the movie titles with “jedi” in their name have jumped to the top of the list, and their rank has increased.
请注意，名称中带有“绝地武士”的电影标题如何跃居列表顶部，并且他们的排名有所提高。

It's worth pointing out that having only four weight “classes” is somewhat limiting, and that they need to be applied when computing the `tsvector`.
值得指出的是，只有四个权重“类”是有一定限制的，在计算 . `tsvector`

[#Typo-tolerance / fuzzy search
错别字容错/模糊搜索](https://xata.io/blog/postgres-full-text-search-engine#typo-tolerance-fuzzy-search)

PostgreSQL doesn't support fuzzy search or typo-tolerance directly, when using `tsvector` and `tsquery`. However, working on the assumptions that the typo is in the query part, we can implement the following idea:
PostgreSQL 在使用 `tsvector` 和 `tsquery` 时不直接支持模糊搜索或拼写错误容错。但是，假设拼写错误在查询部分，我们可以实现以下想法：

- index all _lexemes_ from the content in a separate table
  在单独的表中索引内容中的所有词素
- for each word in the query, use similarity or Levenshtein distance to search in this table
  对于查询中的每个单词，请使用相似性或列文施泰因距离在此表中搜索
- modify the query to include any words that are found
  修改查询以包含找到的任何单词
- perform the search 执行搜索

Here is how it works. First, use `ts_stats` to get all words in a materialized view:
这是它的工作原理。首先，用于 `ts_stats` 获取物化视图中的所有单词：

```
CREATE MATERLIAZED VIEW unique_lexeme AS   SELECT word FROM ts_stat('SELECT search FROM movies');
```

Now, for each word in the query, check if it is in the `unique_lexeme` view. If it's not, do a fuzzy-search in that view to find possible misspellings of it:
现在，对于查询中的每个单词，检查它是否在 `unique_lexeme` 视图中。如果不是，请在该视图中进行模糊搜索以查找可能的拼写错误：

```
SELECT * FROM unique_lexeme   WHERE levenshtein_less_equal(word, 'pregant', 2) < 2;    word---------- premant pregrant pregnant paegant
```

In the above we use the [Levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance) because that's what search engines like Elasticsearch use for fuzzy search.
在上面，我们使用 Levenshtein 距离，因为这是像 Elasticsearch 这样的搜索引擎用于模糊搜索的距离。

Once you have the candidate list of words, you need to adjust the query include them all.
获得候选单词列表后，需要调整查询包括所有单词。

[#Faceted search 分面搜索](https://xata.io/blog/postgres-full-text-search-engine#faceted-search)

Faceted search is popular especially on e-commerce sites because it helps customers to iteratively narrow their search. Here is an example from amazon.com:
分面搜索尤其在电子商务网站上很受欢迎，因为它可以帮助客户以迭代方式缩小搜索范围。以下是 amazon.com 的示例：

[![Faceted search on Amazon](https://xata.io/_next/image?url=%2Fmdx%2Fblog%2Famazon_faceted_search.png&w=3840&q=75)](https://xata.io/mdx/blog/amazon_faceted_search.png)

Faceted search on Amazon 亚马逊上的分面搜索

The above can implemented by manually defining categories and then adding them as `WHERE` conditions to the search. Another approach is to create the categories algorithmically based on the existing data. For example, you can use the following to create a “Decade” facet:
上述可以通过手动定义类别，然后将其作为条件添加到 `WHERE` 搜索中来实现。另一种方法是基于现有数据以算法方式创建类别。例如，您可以使用以下内容创建“十年”分面：

```
SELECT ReleaseYear/10*10 decade, count(Title) cnt FROM movies  WHERE search @@ websearch_to_tsquery('english','star wars')  GROUP BY decade ORDER BY cnt DESC; decade | cnt--------+-----   2000 |  39   2010 |  31   1990 |  29   1950 |  28   1940 |  26   1980 |  22   1930 |  13   1960 |  11   1970 |   7   1910 |   3   1920 |   3(11 rows)
```

This also provides counts of matches for each decade, which you can display in brackets.
这还提供了每个十年的匹配计数，您可以在括号中显示。

If you want to get multiple facets in a single query, you can combine them, for example, by using CTEs:
如果要在单个查询中获取多个方面，可以组合它们，例如，通过使用 CTE：

```sql
WITH releaseYearFacets AS (
  SELECT 'Decade' facet, (ReleaseYear/10*10)::text val, count(Title) cnt
  FROM movies
  WHERE search @@ websearch_to_tsquery('english','star wars')
  GROUP BY val ORDER BY cnt DESC),
genreFacets AS (
  SELECT 'Genre' facet, Genre val, count(Title) cnt FROM movies
  WHERE search @@ websearch_to_tsquery('english','star wars')
  GROUP BY val ORDER BY cnt DESC LIMIT 5)
SELECT * FROM releaseYearFacets UNION SELECT * FROM genreFacets;

 facet  |   val   | cnt
--------+---------+-----
 Decade | 1910    |   3
 Decade | 1920    |   3
 Decade | 1930    |  13
 Decade | 1940    |  26
 Decade | 1950    |  28
 Decade | 1960    |  11
 Decade | 1970    |   7
 Decade | 1980    |  22
 Decade | 1990    |  29
 Decade | 2000    |  39
 Decade | 2010    |  31
 Genre  | comedy  |  21
 Genre  | drama   |  35
 Genre  | musical |   9
 Genre  | unknown |  13
 Genre  | war     |  15
(16 rows)
```

The above should work quite well on small to medium data sets, however it can become slow on very large data sets.
上述方法应该在中小型数据集上工作得很好，但是在非常大的数据集上可能会变慢。

[#Conclusion 结论](https://xata.io/blog/postgres-full-text-search-engine#conclusion)

We've seen the PostgreSQL full-text search primitives, and how we can combine them to create a pretty advanced full-text search engine, which also happens to support things like joins and ACID transactions. In other words, it has functionality that the other search engines typically don't have.
我们已经看到了 PostgreSQL 全文搜索原语，以及如何将它们组合起来创建一个非常高级的全文搜索引擎，它也恰好支持 joins 和 ACID 事务等功能。换句话说，它具有其他搜索引擎通常没有的功能。

There are more advanced search topics that would be worth covering in detail:
还有更高级的搜索主题值得详细介绍：

- suggesters / auto-complete
  建议器/自动完成
- exact phrase matching 完全匹配的短语
- hybrid search (semantic + keyword) by combining with pg-vector
  通过与 PG 向量相结合的混合搜索（语义+关键字）

Each of these would be worth their own blog post (coming!), but by now you should have an intuitive feeling about them: they are quite possible using PostgreSQL, but they require you to do the work of combining the primitives and in some cases the performance might suffer on very large datasets.
这些中的每一个都值得他们自己的博客文章（来！），但现在你应该对它们有一个直观的感觉：它们很可能使用 PostgreSQL，但它们需要你做组合原语的工作，在某些情况下，性能可能会在非常大的数据集上受到影响。

In part 2, we'll make a detailed comparison with Elasticsearch, looking to answer the question on when is it worth it to implement search into PostgreSQL versus adding Elasticsearch to your infrastructure and syncing the data. If you want to be notified when this gets published, you can follow us on [Twitter](https://twitter.com/xata) or join our [Discord](https://xata.io/discord).
在第 2 部分中，我们将与 Elasticsearch 进行详细比较，希望回答何时值得在 PostgreSQL 中实现搜索与将 Elasticsearch 添加到您的基础架构并同步数据的问题。如果您想在发布时收到通知，您可以在 Twitter 上关注我们或加入我们的 Discord。
