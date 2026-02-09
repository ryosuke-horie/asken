import Testing
@testable import Uchikomi

@Suite struct MeasurementUnitTests {
    // MARK: - 単位定義テスト

    @Test func 全単位が正しく定義されているべき() {
        #expect(MeasurementUnit.allCases.count == 16)
    }

    @Test func グラム単位が正しく定義されているべき() {
        let unit = MeasurementUnit.gram
        #expect(unit.rawValue == "g")
        #expect(unit.displayName == "g")
        #expect(unit.inputType == .decimal)
    }

    @Test func 杯単位が正しく定義されているべき() {
        let unit = MeasurementUnit.cup
        #expect(unit.rawValue == "杯")
        #expect(unit.displayName == "杯")
        #expect(unit.inputType == .integer)
    }

    @Test func 合単位が正しく定義されているべき() {
        let unit = MeasurementUnit.go
        #expect(unit.rawValue == "合")
        #expect(unit.displayName == "合")
        #expect(unit.inputType == .integer)
    }

    // MARK: - InputTypeテスト

    @Test func グラムのみ小数入力が可能なべき() {
        #expect(MeasurementUnit.gram.inputType == .decimal)
        #expect(MeasurementUnit.cup.inputType == .integer)
        #expect(MeasurementUnit.go.inputType == .integer)
    }

    // MARK: - Identifiableテスト

    @Test func idがrawValueと同じであるべき() {
        for unit in MeasurementUnit.allCases {
            #expect(unit.id == unit.rawValue)
        }
    }

    // MARK: - Codableテスト

    @Test func エンコードとデコードが正しく動作するべき() {
        let original = MeasurementUnit.go
        let encoded = try? JSONEncoder().encode(original)
        #expect(encoded != nil)

        let decoded = try? JSONDecoder().decode(MeasurementUnit.self, from: encoded!)
        #expect(decoded == original)
    }
}
