{{> _header}}
你好 {{name|顾客}}，你的订单 {{order_id}} 已发货。
{{#if is_vip}}
尊敬的 VIP 用户，您享受 {{discount}} 优惠。
{{/if}}
感谢您的购买！