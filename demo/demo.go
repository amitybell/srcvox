package demo

import "github.com/amitybell/srcvox/config"

const (
	Username = "* PP * FPS DOUG"
	Width    = 600 * 2
	Height   = 500 * 2
)

var (
	Enabled = config.DefaultConfig.Demo != nil && *config.DefaultConfig.Demo

	// stolen from MC:V
	PlayerNames = []string{
		"Charles McKiernan",
		"Hayden Chambers",
		"Nguyễn Ngọc Ẩn",
		"Đỗ Anh Khôi",
		"Hà Tất Hòa",
		"Lâm Khải Tuấn",
		"Ngô Hoàng Nam",
		"Thảo Quang Linh",
		"Triệu Yên Bằng",
		"Carter O'Reilley",
		"Chu Hoàng Minh",
		"Drake Lynch",
		"Huỳnh Minh Chiến",
		"An Anh Quân",
		"Benton Harris",
		"Huỳnh Ðình Luận",
		"Vincent K. Brooks",
		"Leonel Rush",
		"Lưu Ngọc Danh",
		"Patrick Owen",
		"Trầm Huy Thành",
		"Chris Crawford",
		"Kiều Thiện Sinh",
		"Samuel F. Troy",
		"Thân Anh Duy",
		"Bạch Việt Khải",
		"Henry Westmoreland",
		"Lê Ðông Nguyên",
		"Trương Công Tuấn",
		"Đàm Phước Nhân",
		"Nguyễn Ðăng Khánh",
		"Robert McNamara",
		"Martin C. Weyand",
		"Nghiêm Quang Thắng",
		"Christopher Hawkins",
		"Đỗ Minh Hiếu",
		"Lê Trường Vũ",
		"Ryan Grant",
		"Lê Duy Mạnh",
		"Stanley Rose",
		"Bùi Gia Khánh",
		"Châu Chiêu Quân",
		"Úc Tuấn Kiệt",
		"Dylan Watson",
		"Mạc Minh Triết",
		"Steven L. Farrell",
		"Võ Trường Giang",
		"An Thống Nhất",
		"William O'Neal",
		"Huỳnh Anh Vũ",
		"Jake Miller",
		"Blake Webb",
		"Finn Moses",
		"Hoàng Hoàng Giang",
		"Hồ Nghĩa Dũng",
		"Diệp Thái Hòa",
		"Jack Cameron",
		"Lục Thành Trung",
		"Trần Bảo Tín",
		"Vũ Duy Khang",
		"Eric A. Schwartz",
		"Jake Thomson",
		"Lê Triều Thành",
		"Văn Hữu Tài",
		"Finlay Ball",
		"Gary Downing",
		"Ngô Ðức Toàn",
		"Thân Tấn Nam",
		"Benjamin Brown",
		"Trịnh Duy Bảo",
		"Võ Hồng Việt",
		"Xander Montoya",
		"Maddox Pierce",
		"Nguyễn Quốc Mạnh",
		"Phan Vĩnh Toàn",
		"Tommy J. Hendrix",
		"Anthony McCarthy",
		"Patrick Howard",
		"Trang Thụy Miên",
		"Úc Minh Giang",
		"Bùi Mạnh Hà",
		"Bùi Vĩnh Hưng",
		"Diệp Ðức Toàn",
		"Huỳnh Tấn Tài",
		"Sái Trường Nam",
		"Chu Vĩnh Luân",
		"Joseph W. Somervell",
		"Ngô Quang Ninh",
		"Xander Forbes",
		"Daniel Jefferson",
		"Kyree Stokes",
		"Lý Quốc Trụ",
		"Nguyễn Ðình Dương",
		"Heath Bishop",
		"Lesley I. Holmes",
		"Nguyễn Hồng Vinh",
		"Đỗ Quang Trọng",
		"Lê Tuấn Minh",
		"Mark Devers",
		"Trevon Ellis",
		"Đặng Ngọc Lân",
		"Frank P. Harris",
		"Hà Hoàng Duệ",
		"Silas Hays",
		"Võ Khánh Hoàng",
		"Hoàng Minh Kỳ",
		"Lạc Chí Thành",
		"Ngư Khánh Văn",
		"Võ Ðình Phúc",
		"Đàm Thành Doanh",
		"Davon Patel",
		"Gordon Kingston",
		"Nguyễn Huy Hoàng",
		"Trần Phú Hưng",
		"Harry Spencer",
		"Hoàng Minh Tú",
		"Ralph E. Abrams",
		"Trang Thái Tân",
		"Keith McChrystal",
		"Logan Barrett",
		"Phan Huy Vũ",
		"Đỗ Chí Nam",
		"Frederick Davison",
		"Jayden Read",
		"Liễu Hồng Việt",
		"Ngô Quảng Ðạt",
		"Bruce Richardson",
		"Michael Richardson",
		"Trầm Chí Bảo",
		"Vương Duy Cẩn",
		"Dương Minh Dũng",
		"Lê Cao Tiến",
		"Nguyễn Hữu Bình",
		"Quang Quang Tuấn",
		"Văn Chí Công",
		"Nguyễn Long Giang",
		"Paul Wheeler",
		"Payton Russo",
		"Úc Hải Phong",
		"Calvin Bradford",
		"Guy S. Powell",
		"Nguyễn Hải Phong 32",
		"Sái Huy Vũ",
		"Hugh O'Meara",
		"Lucas Barrett",
		"Lý Khắc Vũ",
		"Nguyễn Tuấn Tài",
		"Aydan Wolfe",
		"Dennis Rodriguez",
		"Ngô Vương Triều",
		"Thủy Khởi Phong",
		"Ben T. Porter",
		"Hoàng Trường An",
		"Logan Owen",
		"Lưu Nguyên Hạnh",
		"Bạch Hoàng Vương",
		"Lê Hữu Hạnh",
		"Lạc Chấn Hùng",
		"Sái Anh Quốc",
		"Văn Minh Hiếu",
		"Châu Ðức Cường",
		"Donald P. Green",
		"Ngô Phú Ân",
		"Patrick Lawson",
		"Alex Atkinson",
		"Jonathan Hunt",
		"Nguyễn Nam Sơn",
		"Tiêu Quốc Khánh",
		"Bentley Dale",
		"Harrison Cooke",
		"Lục Long Vịnh",
		"Đặng Trung Nhân",
		"Cao Ðình Toàn",
		"Matthew King",
		"Patrick Lawson",
		"Tôn Minh Nhật",
	}
)