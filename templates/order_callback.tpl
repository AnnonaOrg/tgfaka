订单完成
完成时间:{{TimestampToDatetime .Order.EndTime}}
交易凭证:{{.Order.ID}}
支付金额:{{.Order.Price}} {{.Order.Currency}}-{{.Order.Network}}
商品名称:{{.Product.Name}}
购买数量: {{.BuyNum}}
购买内容:
{{range .ProductItemList}}
- <code>{{.Content}}</code>
{{end}}