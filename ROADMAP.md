# monkey-go Feature Roadmap

このドキュメントは、monkey-goに追加する機能の優先順位付けされたロードマップです。

## 🎯 設計目標

1. **チューリング完全性の実用的実現** - 再帰だけでなく、実用的なイテレーション構造
2. **型安全性** - より安全で予測可能なコード実行
3. **書きやすさ** - 直感的で表現力の高い構文

## 📝 現在の実装状況

### ✅ 実装済み（本の範囲）

- **データ型**: 整数、真偽値、文字列、配列、ハッシュマップ
- **制御構造**: if-else式
- **関数**: クロージャをサポートする第一級関数
- **演算子**: 算術（+, -, *, /）、比較（<, >, ==, !=）、否定（!）
- **組み込み関数**: len, first, last, rest, push, print
- **その他**: Unicode完全サポート

### ⚠️ 本では実装されていない機能

「Writing An Interpreter In Go」の公式実装では、ループ構造と論理演算子は含まれていません。
Monkey言語は元々、再帰と高階関数でイテレーションを実現する設計です。

---

## 🚀 機能追加ロードマップ

### Phase 1: チューリング完全性の実用化 🔴 **最優先**

これらの機能により、理論的なチューリング完全性を**実用的な**チューリング完全性に変換します。

#### 1.1 変数の再代入
**現状**: `let`で束縛後は変更不可
**目標**: 変数の値を更新可能に

```monkey
let x = 5;
x = 10;  // 現在はエラー → 可能にする

let counter = 0;
while (counter < 10) {
    counter = counter + 1;  // ループカウンタに必須
}
```

**実装箇所**:
- `object/environment.go`: 環境での値更新ロジック
- `evaluator/evaluator.go`: 代入文の評価

**優先度**: 🔴 最高
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

#### 1.2 論理演算子（&&, ||） ⭐ READMEのTODO
**現状**: `!`のみサポート
**目標**: AND、ORの論理演算を追加

```monkey
if (x > 0 && x < 10) {
    print("x is between 0 and 10");
}

if (isValid || isAdmin) {
    print("Access granted");
}
```

**実装箇所**:
- `token/token.go`: `AND`、`OR`トークン追加
- `lexer/lexer.go`: `&&`、`||`の字句解析
- `parser/parser.go`: 中置演算子パース
- `evaluator/evaluator.go`: 短絡評価の実装

**優先度**: 🔴 最高
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

#### 1.3 whileループ
**現状**: ループ構造なし（再帰のみ）
**目標**: while文の追加

```monkey
let i = 0;
while (i < 10) {
    print(i);
    i = i + 1;
}
```

**実装箇所**:
- `token/token.go`: `WHILE`キーワード追加
- `ast/ast.go`: `WhileStatement`ノード追加
- `parser/parser.go`: while文のパース
- `evaluator/evaluator.go`: ループ評価

**優先度**: 🔴 最高
**難易度**: ⭐⭐⭐ 高
**依存関係**: 変数の再代入（1.1）

---

#### 1.4 forループ ⭐ READMEのTODO
**現状**: ループ構造なし
**目標**: 3種類のforループ

```monkey
// C-style for
for (let i = 0; i < 10; i = i + 1) {
    print(i);
}

// for-in (イテレータ)
for (item in array) {
    print(item);
}

// 範囲ベース（Phase 3で範囲型実装後）
for (i in 0..10) {
    print(i);
}
```

**実装箇所**:
- `token/token.go`: `FOR`, `IN`キーワード
- `ast/ast.go`: `ForStatement`ノード
- `parser/parser.go`: for文のパース
- `evaluator/evaluator.go`: ループ評価

**優先度**: 🔴 最高
**難易度**: ⭐⭐⭐⭐ 高
**依存関係**: 変数の再代入（1.1）

---

#### 1.5 break/continue文
**現状**: なし
**目標**: ループの制御フロー

```monkey
for (let i = 0; i < 10; i = i + 1) {
    if (i == 5) {
        continue;  // スキップ
    }
    if (i == 8) {
        break;     // ループ脱出
    }
    print(i);
}
```

**実装箇所**:
- `token/token.go`: `BREAK`, `CONTINUE`キーワード
- `ast/ast.go`: `BreakStatement`, `ContinueStatement`
- `evaluator/evaluator.go`: 特殊な戻り値で制御フロー実装

**優先度**: 🟠 高
**難易度**: ⭐⭐⭐ 高
**依存関係**: while/forループ（1.3, 1.4）

---

### Phase 2: 基本的な型安全性 🟠

#### 2.1 実行時型チェックの強化
**現状**: 基本的な型エラー検出
**目標**: 詳細なエラーメッセージ

```monkey
let x = 5 + "hello";
// Before: ERROR: type mismatch: INTEGER + STRING
// After:  ERROR: Type mismatch at line 1, column 9
//         Cannot add INTEGER (5) and STRING ("hello")
//         Suggestion: Convert types or use string concatenation
```

**実装箇所**:
- `evaluator/evaluator.go`: エラーメッセージの改善
- `ast/ast.go`: 行番号・列番号の追跡
- `lexer/lexer.go`: 位置情報の記録

**優先度**: 🟠 高
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

#### 2.2 オプショナル型アノテーション
**現状**: 動的型付けのみ
**目標**: 任意の型アノテーション

```monkey
// 型アノテーション付き
let add: fn(int, int) -> int = fn(x: int, y: int): int {
    return x + y;
};

// 型なし（従来通り）
let add = fn(x, y) { x + y };
```

**実装箇所**:
- `token/token.go`: 型関連トークン（`:`, `->`）
- `ast/ast.go`: 型アノテーションノード
- `parser/parser.go`: 型構文のパース
- 新規: `typechecker/`: 型チェッカーパッケージ

**優先度**: 🟡 中
**難易度**: ⭐⭐⭐⭐⭐ 非常に高
**依存関係**: Phase 1完了

---

#### 2.3 Null安全性（Option型）
**現状**: `null`は暗黙的
**目標**: 明示的なOption型

```monkey
// Option型の導入
let result: Option<int> = Some(42);
let empty: Option<int> = None;

// パターンマッチング（Phase 3で実装）
match result {
    Some(value) => print(value),
    None => print("No value")
}

// Null合体演算子（Phase 3）
let value = maybeNull ?? defaultValue;
```

**実装箇所**:
- `object/object.go`: `Option`オブジェクト型
- `evaluator/evaluator.go`: Option操作

**優先度**: 🟡 中
**難易度**: ⭐⭐⭐⭐ 高
**依存関係**: 型アノテーション（2.2）、パターンマッチング（3.5）

---

#### 2.4 Result型とエラーハンドリング
**現状**: エラーオブジェクトのみ
**目標**: Result型による明示的エラー処理

```monkey
fn divide(a: int, b: int): Result<int, string> {
    if (b == 0) {
        return Err("Division by zero");
    }
    return Ok(a / b);
}

let result = divide(10, 0);
match result {
    Ok(value) => print("Result: " + value),
    Err(error) => print("Error: " + error)
}
```

**実装箇所**:
- `object/object.go`: `Result`オブジェクト型
- `evaluator/evaluator.go`: Result操作

**優先度**: 🟡 中
**難易度**: ⭐⭐⭐⭐ 高
**依存関係**: 型アノテーション（2.2）、パターンマッチング（3.5）

---

### Phase 3: 書きやすさの向上 🟡

#### 3.1 比較演算子の完全化
**現状**: `<`, `>`, `==`, `!=`のみ
**目標**: `>=`, `<=`を追加

```monkey
if (x >= 5 && y <= 10) {
    print("In range");
}
```

**実装箇所**:
- `token/token.go`: `GTE`, `LTE`トークン
- `lexer/lexer.go`: `>=`, `<=`の認識
- `parser/parser.go`: 演算子パース
- `evaluator/evaluator.go`: 評価ロジック

**優先度**: 🟠 高
**難易度**: ⭐ 低
**依存関係**: なし

---

#### 3.2 算術演算子の拡充
**現状**: +, -, *, / のみ
**目標**: 剰余、累乗、複合代入、インクリメント

```monkey
let remainder = 10 % 3;      // 剰余
let power = 2 ** 8;          // 累乗
x += 5;                      // 複合代入
x++;                         // インクリメント
x--;                         // デクリメント
```

**実装箇所**:
- `token/token.go`: 新しい演算子トークン
- `lexer/lexer.go`: 演算子認識
- `parser/parser.go`: パース
- `evaluator/evaluator.go`: 評価

**優先度**: 🟠 高
**難易度**: ⭐⭐ 中
**依存関係**: 変数の再代入（1.1 - 複合代入用）

---

#### 3.3 浮動小数点数サポート
**現状**: 整数のみ
**目標**: float型の追加

```monkey
let pi = 3.14159;
let result = 10.5 / 2.0;
let x = 5;      // int
let y = 5.0;    // float
```

**実装箇所**:
- `token/token.go`: `FLOAT`トークン
- `lexer/lexer.go`: 浮動小数点リテラル
- `object/object.go`: `Float`オブジェクト
- `evaluator/evaluator.go`: 浮動小数点演算

**優先度**: 🟡 中
**難易度**: ⭐⭐⭐ 高
**依存関係**: なし

---

#### 3.4 範囲型
**現状**: なし
**目標**: 範囲リテラルとイテレーション

```monkey
let numbers = 1..10;           // [1,2,3,...,10]
let evens = 0..100 step 2;     // [0,2,4,...,100]

for (i in 1..10) {
    print(i);
}
```

**実装箇所**:
- `token/token.go`: `RANGE`トークン (`..`)
- `ast/ast.go`: `RangeExpression`
- `object/object.go`: `Range`オブジェクト
- `evaluator/evaluator.go`: 範囲評価

**優先度**: 🟢 低
**難易度**: ⭐⭐⭐ 高
**依存関係**: forループ（1.4）

---

#### 3.5 パターンマッチング
**現状**: if-else式のみ
**目標**: match式による強力なパターンマッチング

```monkey
match value {
    0 => print("zero"),
    1..10 => print("small number"),
    [first, ...rest] => print("array"),
    {name: n, age: a} => print(n + " is " + a),
    _ => print("default")
}
```

**実装箇所**:
- `token/token.go`: `MATCH`, `ARROW`（`=>`）
- `ast/ast.go`: `MatchExpression`, `MatchArm`
- `parser/parser.go`: match構文パース
- `evaluator/evaluator.go`: パターン評価

**優先度**: 🟡 中
**難易度**: ⭐⭐⭐⭐⭐ 非常に高
**依存関係**: Phase 1完了

---

#### 3.6 デストラクチャリング
**現状**: なし
**目標**: 配列とハッシュマップの分解代入

```monkey
let [a, b, ...rest] = [1, 2, 3, 4, 5];
let {name, age} = person;
let {name: userName, age: userAge} = person;
```

**実装箇所**:
- `ast/ast.go`: デストラクチャパターン
- `parser/parser.go`: 分解代入パース
- `evaluator/evaluator.go`: パターン分解

**優先度**: 🟢 低
**難易度**: ⭐⭐⭐⭐ 高
**依存関係**: パターンマッチング（3.5）

---

#### 3.7 スプレッド演算子
**現状**: なし
**目標**: 配列とハッシュマップの展開

```monkey
let arr1 = [1, 2, 3];
let arr2 = [...arr1, 4, 5, 6];
let merged = {...obj1, ...obj2};
```

**実装箇所**:
- `token/token.go`: `SPREAD`トークン (`...`)
- `parser/parser.go`: スプレッド構文
- `evaluator/evaluator.go`: 配列・ハッシュ展開

**優先度**: 🟢 低
**難易度**: ⭐⭐⭐ 高
**依存関係**: なし

---

#### 3.8 文字列補間/テンプレート
**現状**: 文字列連結のみ
**目標**: テンプレート文字列

```monkey
let name = "World";
let greeting = `Hello, ${name}!`;
let multiline = `Line 1
Line 2
Line 3`;
```

**実装箇所**:
- `token/token.go`: テンプレート文字列トークン
- `lexer/lexer.go`: バッククォート処理
- `parser/parser.go`: 補間式パース
- `evaluator/evaluator.go`: テンプレート評価

**優先度**: 🟢 低
**難易度**: ⭐⭐⭐ 高
**依存関係**: なし

---

#### 3.9 コメントサポート（複数行）
**現状**: 不明（単一行コメントの有無を確認必要）
**目標**: 複数行コメント

```monkey
// 単一行コメント

/*
 * 複数行コメント
 * ブロックコメント
 */
```

**実装箇所**:
- `lexer/lexer.go`: コメント処理（トークンとしては無視）

**優先度**: 🟢 低
**難易度**: ⭐ 低
**依存関係**: なし

---

### Phase 4: 標準ライブラリの拡充 🟢

#### 4.1 配列操作関数
**現状**: len, first, last, rest, push
**目標**: 高階関数とユーティリティ

```monkey
// 高階関数
let doubled = [1,2,3].map(fn(x) { x * 2 });
let evens = [1,2,3,4].filter(fn(x) { x % 2 == 0 });
let sum = [1,2,3].reduce(fn(acc, x) { acc + x }, 0);

// ユーティリティ
let found = [1,2,3].find(fn(x) { x > 2 });
let hasEven = [1,2,3].some(fn(x) { x % 2 == 0 });
let allPositive = [1,2,3].every(fn(x) { x > 0 });

// 変更系
let popped = arr.pop();
let reversed = arr.reverse();
let sorted = arr.sort(fn(a, b) { a - b });
```

**実装箇所**:
- `evaluator/builtins.go`: 新しい組み込み関数

**優先度**: 🟡 中
**難易度**: ⭐⭐ 中
**依存関係**: Phase 1完了

---

#### 4.2 文字列操作関数
**現状**: 文字列連結のみ
**目標**: 豊富な文字列操作

```monkey
let parts = "a,b,c".split(",");           // ["a", "b", "c"]
let joined = ["a", "b", "c"].join("-");   // "a-b-c"
let replaced = "hello".replace("l", "L"); // "heLLo"
let sub = "hello".substring(0, 2);        // "he"
let upper = "hello".toUpperCase();        // "HELLO"
let lower = "HELLO".toLowerCase();        // "hello"
let trimmed = "  hello  ".trim();         // "hello"
```

**実装箇所**:
- `evaluator/builtins.go`: 文字列関数追加

**優先度**: 🟡 中
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

#### 4.3 数学関数
**現状**: 基本演算のみ
**目標**: 標準的な数学関数

```monkey
let absolute = abs(-5);        // 5
let rounded = round(3.7);      // 4
let floored = floor(3.7);      // 3
let ceiled = ceil(3.2);        // 4
let squareRoot = sqrt(16);     // 4
let power = pow(2, 3);         // 8
let minimum = min(1, 2, 3);    // 1
let maximum = max(1, 2, 3);    // 3
let random = random();         // 0.0 ~ 1.0
```

**実装箇所**:
- `evaluator/builtins.go`: 数学関数追加

**優先度**: 🟡 中
**難易度**: ⭐ 低
**依存関係**: 浮動小数点数（3.3）

---

#### 4.4 型操作関数
**現状**: 実行時型エラーのみ
**目標**: 型の検査と変換

```monkey
let t = type(42);              // "INTEGER"
let isInt = typeof(x, "int");  // true/false

let num = parseInt("123");     // 123
let float = parseFloat("3.14"); // 3.14
let str = toString(42);        // "42"
```

**実装箇所**:
- `evaluator/builtins.go`: 型関数追加

**優先度**: 🟢 低
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

#### 4.5 I/O関数
**現状**: print のみ
**目標**: 入力と整形出力

```monkey
let name = input("Enter your name: ");
print("Hello, " + name);

printf("Name: %s, Age: %d\n", name, age);
let formatted = sprintf("Value: %d", 42);
```

**実装箇所**:
- `evaluator/builtins.go`: I/O関数追加

**優先度**: 🟢 低
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

#### 4.6 ファイルI/O
**現状**: なし
**目標**: ファイル読み書き

```monkey
let content = readFile("input.txt");
writeFile("output.txt", "Hello, World!");

let lines = readLines("data.txt");
let exists = fileExists("test.txt");
```

**実装箇所**:
- `evaluator/builtins.go`: ファイルI/O関数

**優先度**: 🟢 低
**難易度**: ⭐⭐ 中
**依存関係**: なし

---

### Phase 5: モジュールシステム 🔵

#### 5.1 import/export
**現状**: なし
**目標**: モジュールシステム

```monkey
// math.monkey
export fn add(x, y) {
    return x + y;
}

export let PI = 3.14159;

// main.monkey
import { add, PI } from "./math";
import * as math from "./math";

print(add(1, 2));
print(math.PI);
```

**実装箇所**:
- `token/token.go`: `IMPORT`, `EXPORT`キーワード
- `ast/ast.go`: import/exportノード
- `parser/parser.go`: モジュール構文
- `evaluator/evaluator.go`: モジュール解決
- 新規: `module/`: モジュールシステム

**優先度**: 🔵 将来的
**難易度**: ⭐⭐⭐⭐⭐ 非常に高
**依存関係**: Phase 1-3完了

---

### Phase 6: 高度な型システム 🔵

#### 6.1 型推論
**現状**: 動的型付け
**目標**: 自動型推論

```monkey
let x = 5;           // 推論: int
let y = x + 3;       // 推論: int
let z = "hello";     // 推論: string
```

**実装箇所**:
- 新規: `typechecker/inference.go`: 型推論エンジン

**優先度**: 🔵 将来的
**難易度**: ⭐⭐⭐⭐⭐ 非常に高
**依存関係**: 型アノテーション（2.2）

---

#### 6.2 ジェネリクス
**現状**: なし
**目標**: パラメトリック多相

```monkey
fn identity<T>(x: T): T {
    return x;
}

fn map<T, U>(arr: Array<T>, f: fn(T) -> U): Array<U> {
    // ...
}
```

**実装箇所**:
- `ast/ast.go`: ジェネリック構文
- `typechecker/`: 型パラメータ解決

**優先度**: 🔵 将来的
**難易度**: ⭐⭐⭐⭐⭐ 非常に高
**依存関係**: 型推論（6.1）

---

### Phase 7: コンパイラへの進化 ⭐

#### 7.1 「Writing A Compiler In Go」の実装
**現状**: ツリーウォーキングインタプリタ
**目標**: バイトコードコンパイラ + VM

- バイトコード命令セットの定義
- コンパイラ（AST → バイトコード）
- スタックベースVM
- シンボルテーブルと定数プール
- **約3倍のパフォーマンス向上**

**実装箇所**:
- 新規: `compiler/`: コンパイラパッケージ
- 新規: `vm/`: 仮想マシン
- 新規: `code/`: バイトコード定義

**優先度**: ⭐ 長期目標
**難易度**: ⭐⭐⭐⭐⭐ 非常に高
**依存関係**: Phase 1-3完了推奨

---

## 📊 推奨実装順序

### 🎯 Quick Wins（短期 - 1〜2週間）

実用性が高く、実装が比較的容易な機能から着手：

1. **論理演算子（&&, ||）** - 1〜2日
2. **比較演算子（>=, <=）** - 1日
3. **剰余演算子（%）** - 1日
4. **変数の再代入** - 2〜3日

→ **これだけで実用性が大幅に向上！**

---

### 🏃 Short Term（中期 - 1〜2ヶ月）

**Phase 1を完了させる：**

5. **whileループ** - 3〜5日
6. **forループ** - 5〜7日
7. **break/continue** - 2〜3日
8. **複合代入演算子** - 2日
9. **浮動小数点数** - 5〜7日

→ **実用的なチューリング完全言語になる**

---

### 🚶 Mid Term（長期 - 3〜6ヶ月）

**Phase 2 & 3を進める：**

10. エラーメッセージ改善
11. 標準ライブラリ拡充（配列・文字列）
12. オプショナル型アノテーション
13. パターンマッチング

→ **型安全性と開発体験が向上**

---

### 🏔️ Long Term（超長期 - 6ヶ月以上）

**Phase 4-7を検討：**

14. モジュールシステム
15. 型推論・ジェネリクス
16. コンパイラ版実装

→ **本格的なプログラミング言語に進化**

---

## 💡 実装ガイドライン

### 一般原則

1. **テストファースト**: 各機能に対して包括的なテストを書く
2. **後方互換性**: 既存のコードを壊さない
3. **段階的実装**: 大きな機能は小さなステップに分割
4. **ドキュメント**: 新機能のドキュメントと例を追加

### テストカバレッジ

- 現在の高いカバレッジ（codecov）を維持
- 各新機能で90%以上のカバレッジを目指す
- エッジケースのテストを忘れずに

### コードスタイル

- 既存のGoコーディングスタイルに従う
- golintとgofmtを使用
- CI/CDパイプラインですべてのチェックをパス

---

## 📚 参考資料

### 本家・公式

- [Writing An Interpreter In Go](https://interpreterbook.com/)
- [Writing A Compiler In Go](https://compilerbook.com/)
- [Monkey Language](https://monkeylang.org/)
- [公式コード](https://interpreterbook.com/waiig_code_1.7.zip)

### 拡張実装の参考

- [skx/monkey](https://github.com/skx/monkey) - 豊富な拡張機能
- [bradford-hamilton/monkey-lang](https://github.com/bradford-hamilton/monkey-lang) - コンパイラ版含む
- [prologic/monkey-lang](https://git.mills.io/prologic/monkey-lang) - ステップバイステップ実装

---

## 🤝 貢献

このロードマップは進化します。新しいアイデアや優先順位の変更があれば、
プルリクエストやイシューで提案してください！

---

_最終更新: 2025-11-07_
