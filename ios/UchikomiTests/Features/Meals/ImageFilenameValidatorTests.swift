import Testing
@testable import Uchikomi

@Suite struct ImageFilenameValidatorTests {
    // MARK: - 正常系

    @Test func UUID形式のjpgファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("550e8400-e29b-41d4-a716-446655440000.jpg")
        #expect(result == true)
    }

    @Test func UUID形式のjpegファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("550e8400-e29b-41d4-a716-446655440000.jpeg")
        #expect(result == true)
    }

    @Test func UUID形式のpngファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("550e8400-e29b-41d4-a716-446655440000.png")
        #expect(result == true)
    }

    @Test func UUID形式のheicファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("550e8400-e29b-41d4-a716-446655440000.heic")
        #expect(result == true)
    }

    @Test func 大文字拡張子のファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("image.JPG")
        #expect(result == true)
    }

    @Test func アンダースコア含むファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("my_image_01.png")
        #expect(result == true)
    }

    @Test func ハイフン含むファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("my-image-01.jpg")
        #expect(result == true)
    }

    @Test func 複数ドット含む正当なファイル名はバリデーションを通過すべき() {
        let result = ImageFilenameValidator.isValid("image.backup.jpg")
        #expect(result == true)
    }

    // MARK: - 異常系

    @Test func 空文字列はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("")
        #expect(result == false)
    }

    @Test func パストラバーサルはバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("../../etc/passwd")
        #expect(result == false)
    }

    @Test func ドットドットのみはバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("..")
        #expect(result == false)
    }

    @Test func スラッシュ含むファイル名はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("uploads/image.jpg")
        #expect(result == false)
    }

    @Test func スペース含むファイル名はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("my image.jpg")
        #expect(result == false)
    }

    @Test func gif拡張子はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("image.gif")
        #expect(result == false)
    }

    @Test func svg拡張子はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("image.svg")
        #expect(result == false)
    }

    @Test func 拡張子なしはバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("image")
        #expect(result == false)
    }

    @Test func 特殊文字含むファイル名はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("<script>.jpg")
        #expect(result == false)
    }

    @Test func バックスラッシュ含むファイル名はバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid("..\\etc\\passwd.jpg")
        #expect(result == false)
    }

    @Test func ドット始まりの隠しファイルはバリデーション失敗すべき() {
        let result = ImageFilenameValidator.isValid(".hidden.jpg")
        #expect(result == false)
    }
}
