package iplocation

func NewBuiltinDB() *LocationDB {
	db := NewLocationDB()

	ranges := []struct {
		startIP  string
		endIP    string
		location Location
	}{
		{"1.1.0.0", "1.1.255.255", Location{Country: "中国", Province: "广东省", City: "广州市"}},
		{"14.0.0.0", "14.255.255.255", Location{Country: "中国", Province: "广东省", City: "深圳市"}},
		{"36.0.0.0", "36.63.255.255", Location{Country: "中国", Province: "浙江省", City: "杭州市"}},
		{"36.96.0.0", "36.127.255.255", Location{Country: "中国", Province: "江苏省", City: "南京市"}},
		{"58.240.0.0", "58.255.255.255", Location{Country: "中国", Province: "江苏省", City: "苏州市"}},
		{"59.151.0.0", "59.151.255.255", Location{Country: "中国", Province: "北京市", City: "北京市"}},
		{"101.0.0.0", "101.127.255.255", Location{Country: "中国", Province: "上海市", City: "上海市"}},
		{"111.0.0.0", "111.63.255.255", Location{Country: "中国", Province: "广东省", City: "广州市"}},
		{"112.0.0.0", "112.31.255.255", Location{Country: "中国", Province: "浙江省", City: "杭州市"}},
		{"114.64.0.0", "114.127.255.255", Location{Country: "中国", Province: "广东省", City: "深圳市"}},
		{"116.0.0.0", "116.63.255.255", Location{Country: "中国", Province: "北京市", City: "北京市"}},
		{"116.224.0.0", "116.255.255.255", Location{Country: "中国", Province: "上海市", City: "上海市"}},
		{"120.0.0.0", "120.63.255.255", Location{Country: "中国", Province: "浙江省", City: "宁波市"}},
		{"121.0.0.0", "121.63.255.255", Location{Country: "中国", Province: "上海市", City: "上海市"}},
		{"122.0.0.0", "122.63.255.255", Location{Country: "中国", Province: "江苏省", City: "南京市"}},
		{"123.0.0.0", "123.63.255.255", Location{Country: "中国", Province: "浙江省", City: "杭州市"}},
		{"124.64.0.0", "124.127.255.255", Location{Country: "中国", Province: "北京市", City: "北京市"}},
		{"180.0.0.0", "180.63.255.255", Location{Country: "中国", Province: "广东省", City: "广州市"}},
		{"180.96.0.0", "180.127.255.255", Location{Country: "中国", Province: "江苏省", City: "南京市"}},
		{"183.0.0.0", "183.63.255.255", Location{Country: "中国", Province: "广东省", City: "深圳市"}},
		{"202.96.0.0", "202.96.255.255", Location{Country: "中国", Province: "上海市", City: "上海市"}},
		{"218.0.0.0", "218.63.255.255", Location{Country: "中国", Province: "浙江省", City: "杭州市"}},
		{"219.0.0.0", "219.63.255.255", Location{Country: "中国", Province: "广东省", City: "广州市"}},
		{"220.0.0.0", "220.63.255.255", Location{Country: "中国", Province: "北京市", City: "北京市"}},
		{"221.0.0.0", "221.63.255.255", Location{Country: "中国", Province: "江苏省", City: "南京市"}},
		{"222.0.0.0", "222.63.255.255", Location{Country: "中国", Province: "浙江省", City: "宁波市"}},
	}

	for _, r := range ranges {
		_ = db.AddRange(r.startIP, r.endIP, r.location)
	}

	return db
}
