订单完成
完成时间:{{TimestampToDatetime .Order.EndTime}}
交易凭证:{{.Order.ID}}
支付金额:{{.Order.Price}} {{.Order.Currency}}-{{.Order.Network}}
商品名称:{{.Product.Name}}
一次性邀请入群地址:
{{range .ProductItemList}}
 {{.}} 
{{end}}