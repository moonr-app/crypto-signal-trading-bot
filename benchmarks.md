Results from running: go test -run=xxx  -bench=. -benchtime=20s

|Region     | Binance                | Binance CZ             | 
|-----------|------------------------|------------------------|
| eu-west-1 |204513417 ns/op (204 ms)|366986041 ns/op (366 ms)|
| ap-northeast-1a (Tokyo) |  8146469 ns/op (8.146469 ms) |  175669842 ns/op(175 ms)|
| ap-northeast-2 (Seoul)  | 37393149 ns/op (37.3 ms) | 152504828 ns/op (152ms)|
| ap-northeast-3 (Osaka)  | 14408107 ns/op (14.4 ms)| 174477396 ns/op (174 ms) |
| ap-southeast-1 (Singapore) |  85250410 ns/op (85 ms)|  151169196 ns/op (151 ms) |
| ca-central-1 (Canada) | 170025150 ns/op (107ms)| 309946289 ns/op (309 ms) |
| eu-central-1 (Frankfurt) | 234544604 ns/op (234 ms) | 323968465 ns/op (309 ms)|