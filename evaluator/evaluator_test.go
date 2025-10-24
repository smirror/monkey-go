package evaluator

import (
	"monkey-go/lexer"
	"monkey-go/object"
	"monkey-go/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if (10 > 1) { return 10; }", 10},
		{
			`
		if (10 > 1) {
		 if (10 > 1) {
		   return 10;
		 }
		
		 return 1;
		}
		`,
			10,
		},
		//		{
		//			`
		//let f = fn(x) {
		//  return x;
		//  x + 10;
		//};
		//f(10);`,
		//			10,
		//		},
		//		{
		//			`
		//let f = fn(x) {
		//   let result = x + 10;
		//   return result;
		//   return 10;
		//};
		//f(10);`,
		//			20,
		//		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"true + false + true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
if (10 > 1) {
  if (10 > 1) {
    return true + false;
  }

  return 1;
}
`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, evaluated interface{}, expected int64) bool {
	result, ok := evaluated.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", evaluated, evaluated)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}

	return true
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t",
			result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
let newAdder = fn(x){
	fn(y) { x - y };
};

let addTwo=newAdder(1);
addTwo(2);`

	testIntegerObject(t, testEval(input), -1)
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "世界!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello 世界!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello 世界")`, 8},
		{`len("lorem ipsum")`, 11}, // filler text
		{`len("다람쥐 헌 쳇바퀴에 타고파")`, 14},                                                                                      // korean pangrams
		{`len("Cwm fjord veg balks nth pyx quiz.")`, 33},                                                                   // perfect pangrams
		{`len("Ξεσκεπάζω την ψυχοφθόρα βδελυγμία")`, 33},                                                                   // greek pangrams
		{`len("עטלף אבק נס דרך מזגן שהתפוצץ כי חם")`, 34},                                                                  // hebrew pangrams
		{`len("키스의 고유조건은 입술끼리 만나야 하고 특별한 기술은 필요치 않다")`, 36},                                                                // korean pangrams
		{`len("Ξεσκεπάζω την ψυχοφθόρα βδελυγμία")`, 33},                                                                   // greek pangrams
		{`len("Nyx' Bö drückt Vamps Quiz-Floß jäh weg")`, 38},                                                              // german pangram
		{`len("The quick brown fox jumps over the lazy dog")`, 43},                                                         // pangram
		{`len("Эй, жлоб! Где туз? Прячь юных съёмщиц в шкаф.")`, 45},                                                       // russian pangram                                                                                   // spanish pangram
		{`len("よむほまれをえ きみへちゆうもく ひんとおそわりふやせ めいろぬけて あらのにたね はなさかすこつしる")`, 51},                                                 // japanese pangram
		{`len("とりなくこゑす ゆめさませ みよあけわたる ひんかしを そらいろはえて おきつへに ほふねむれゐぬ もやのうち")`, 55},                                             // japanese pangram
		{`len("Quel vituperabile xenofobo zelante assaggia il whisky ed esclama: alleluja!")`, 75},                         // italian pangram
		{`len("El pingüino Wenceslao hizo kilómetros bajo exhaustiva lluvia y frío, añoraba a su querido cachorro.")`, 99}, // spanish pangram
		{`len("သီဟိုဠ်မှ ဉာဏ်ကြီးရှင်သည် အာယုဝဍ္ဎနဆေးညွှန်းစာကို ဇလွန်ဈေးဘေး ဗာဒံပင်ထက် အဓိဋ္ဌာန်လျက် ဂဃနဏဖတ်ခဲ့သည်။")`, 101},                    // burmese pangrams
		{`len("در صورت حذف این چند واژه غلط به شکیل، ثابت و جامع‌تر ساختن پاراگراف شعر از لحاظ دوری از قافیه‌های اضافه کمک می‌شود")`, 114},       // arabic pangram
		{`len("Dès Noël, où un zéphyr haï me vêt de glaçons würmiens, je dîne d’exquis rôtis de bœuf au kir, à l’aÿ d’âge mûr, &cætera.")`, 120}, // french pangram
		{`len("เป็นมนุษย์สุดประเสริฐเลิศคุณค่า กว่าบรรดาฝูงสัตว์เดรัจฉาน จงฝ่าฟันพัฒนาวิชาการ อย่าล้างผลาญฤๅเข่นฆ่าบีฑาใคร ไม่ถือโทษโกรธแช่งซัดฮึดฮัดด่า หัดอภัยเหมือนกีฬาอัชฌาสัย ปฏิบัติประพฤติกฎกำหนดใจ  ูดจาให้จ๊ะ ๆ จ๋า น่าฟังเอยฯ")`, 216},                                // thai pangrams
		{`len("ঊনিশে কার্তিক রাত্র সাড়ে আট ঘটিকায় ভৈরবনিবাসী ব্যাংকের ক্ষুদ্র ঋণগ্রস্ত অভাবী দুঃস্থ পৌঢ় কৃষক এজাজ মিঞা হাতের কাছে ঔষধ থাকিতেও ঐ ঋণের ডরেই চোখে ঝাপসা দেখিয়া বুকের যন্ত্রণায় ঈষৎ কাঁপিয়া উঠিয়া উঠানে বিছানো ধূসর রঙের ফরাশের উপর ঢলিয়া পড়িলেন।")`, 247}, // bengali pangrams
		{`len("天地玄黃 宇宙洪荒 日月盈昃 辰宿列張 寒來暑往 秋收冬藏 閏餘成歲 律召調陽 雲騰致雨 露結爲霜 金生麗水 玉出崑岡 劍號巨闕 珠稱夜光 果珍李柰 菜重芥薑 海鹹河淡 鱗潛羽翔 龍師火帝 鳥官人皇 始制文字 乃服衣裳 推位讓國 有虞陶唐 弔民伐罪 周發殷湯 坐朝問道 垂拱平章 愛育黎首 臣伏戎羌 遐邇壹體 率賓歸王 鳴鳳在樹 白駒食場 化被草木 賴及萬方 蓋此身髮 四大五常 恭惟鞠養 豈敢毀傷 女慕貞絜 男效才良 知過必改 得能莫忘 罔談彼短 靡恃己長 信使可覆 器欲難量 墨悲絲淬 詩讃羔羊 景行維賢 克念作聖 德建名立 形端表正 空谷傳聲 虛堂習聽 禍因惡積 福緣善慶 尺璧非寶 寸陰是競 資父事君 曰嚴與敬 孝當竭力 忠則盡命 臨深履薄 夙興溫凊 似蘭斯馨 如松之盛 川流不息 淵澄取映 容止若思 言辭安定 篤初誠美 慎終宜令 榮業所基 籍甚無竟 學優登仕 攝職從政 存以甘棠 去而益詠 樂殊貴賤 禮別尊卑 上咊下睦 夫唱婦隨 外受傅訓 入奉母儀 諸姑伯叔 猶子比兒 孔懷兄弟 同气連枝 交友投分 切磨箴規 仁慈隱惻 造次弗離 節義廉退 顛沛匪虧 性靜情逸 心動神疲 守眞志滿 逐物意移 堅持雅操 好爵自縻 都邑華夏 東西二京 背邙面洛 浮渭據涇 宮殿盤鬱 樓觀飛驚 圖寫禽獸 畫彩仙靈 丙舍傍啟 甲帳對楹 肆筵設席 鼓瑟吹笙 升階納陛 弁轉疑星 右通廣內 左達承明 既集墳典 亦聚羣英 杜稾鍾隸 漆書壁經 府羅將相 路俠槐卿 戶封八縣 家給千兵 高冠陪輦 驅轂振纓 世祿侈富 車駕肥輕 策功茂實 勒碑刻銘 磻溪伊尹 佐時阿衡 奄宅曲阜 微旦孰營 桓公匡合 濟弱扶傾 綺迴漢惠 說感武丁 俊乂密勿 多士寔寧 晉楚更霸 趙魏困橫 假途滅虢 踐土會盟 何遵約法 韓弊煩刑 起翦頗牧 用軍最精 宣威沙漠 馳譽丹青 九州禹跡 百郡秦并 嶽宗恆岱 禪主云亭 雁門紫塞 雞田赤城 昆池碣石 鉅野洞庭 曠遠緜邈 巖岫杳冥 治本於農 務茲稼穡 俶載南畝 我藝黍稷 稅熟貢新 勸賞黜陟 孟軻敦素 史魚秉直 庶幾中庸 勞謙謹敕 聆音察理 鑑皃辧色 貽厥嘉猷 勉其祗植 省躬譏誡 寵增抗極 殆辱近恥 林皋幸即 兩疏見機 解組誰逼 索居閒處 沈默寂寥 求古尋論 散慮逍遙 欣奏累遣 慼謝歡招 渠荷的歷 園莽抽條 枇杷晚翠 梧桐早凋 陳根委翳 落葉飄颻 游鯤獨運 夌摩絳霄 耽讀翫市 寓目囊箱 易輶攸畏 屬耳垣牆 具膳喰飯 適口充腸 飽飫亯宰 飢厭糟糠 親戚故舊 老少異糧 妾御績紡 侍巾帷房 紈扇圓潔 銀燭煒煌 晝瞑夕寐 籃筍象牀 弦歌酒讌 接杯舉觴 矯手頓足 悅豫且康 嫡後嗣續 祭祀烝嘗 稽顙再拜 悚懼恐惶 箋牒簡要 顧答審詳 骸垢想浴 執熱願涼 驢騾犢特 駭躍超驤 誅斬賊盜 捕獲叛亡 布射遼丸 嵇琴阮嘯 恬筆倫紙 鈞巧任釣 釋紛利俗 並皆佳妙 毛施淑姿 工顰妍笑 秊矢每催 曦暉朗耀 琁璣懸斡 晦魄環照 指薪脩祜 永綏吉劭 矩步引領 俯仰廊廟 束帶矜莊 徘徊瞻眺 孤陋寡聞 愚蒙等誚 謂語助者 焉哉乎也")`, 1249}, // 千字文
		{`len("🐒")`, 1},
		{`len(1)`, "argument to `len` not supported INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`print("hello", "world!")`, nil},
		{`first([1, 2, 3])`, 1},
		{`first([])`, nil},
		{`first(1)`, "argument to `first` must be ARRAY, got INTEGER"},
		{`last([1, 2, 3])`, 3},
		{`last([])`, nil},
		{`last(1)`, "argument to `last` must be ARRAY, got INTEGER"},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, nil},
		{`push([], 1)`, []int{1}},
		{`push(1, 1)`, "argument to `push` must be ARRAY, got INTEGER"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case nil:
			testNullObject(t, evaluated)
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		case []int:
			array, ok := evaluated.(*object.Array)
			if !ok {
				t.Errorf("obj not Array. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if len(array.Elements) != len(expected) {
				t.Errorf("wrong num of elements. want=%d, got=%d",
					len(expected), len(array.Elements))
				continue
			}

			for i, expectedElem := range expected {
				testIntegerObject(t, array.Elements[i], int64(expectedElem))
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestHashMapLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.HashMap)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashMapIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestAssignmentExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{
			`let x = 5; x = 10; x;`,
			10,
		},
		{
			`let x = 5; x = x + 10; x;`,
			15,
		},
		{
			`let x = 5; let y = x = 10; y;`,
			10,
		},
		{
			`let x = 5; let y = x = 10; x;`,
			10,
		},
		{
			`
			let x = 5;
			let y = 10;
			x = y;
			x;
			`,
			10,
		},
		{
			`
			let x = 5;
			if (true) {
				x = 10;
			}
			x;
			`,
			10,
		},
		{
			`
			let counter = 0;
			let i = 0;
			if (i < 5) {
				counter = counter + 1;
				i = i + 1;
			}
			counter;
			`,
			1,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}
