地址(点击复制):
<code>{{.Order.WalletAddress}}</code>
货币:{{.Order.Currency}}
主网:{{.Order.Network}}
金额:{{.Order.Price}}
{{ .OrderNoteTitle }}{{ .OrderNote }}

创建时间:{{TimestampToDatetime .Order.CreateTime}}
结束时间:{{TimestampToDatetime .Order.EndTime}}

请在规定时间内往上述地址付款指定金额，注意支付的主网类型
支付前请核对图片中的地址和消息中的地址是否一致，确认一致后再进行付款