# Project Glossary (Thuật ngữ dự án)

Bảng dưới đây định nghĩa các thuật ngữ chính được sử dụng xuyên suốt trong mã nguồn và tài liệu của dự án **AI Vocabulary Tutor**.

| Thuật ngữ | Ý nghĩa / Định nghĩa trong ngữ cảnh dự án |
| --- | --- |
| **Spaced Repetition (SRS)** | Hệ thống lặp lại ngắt quãng. Thuật toán lên lịch ôn tập từ vựng dựa trên độ ghi nhớ của người dùng. Từ nào hay quên sẽ xuất hiện nhiều hơn. |
| **Flashcard** | Thẻ từ vựng. Gồm mặt trước (Từ/Câu tiếng Anh) và mặt sau (Định nghĩa/Giải thích). Dùng để kiểm tra khả năng nhớ từ. |
| **Leitner System** | Thuật toán cơ bản của Spaced Repetition, phân loại từ vựng vào các "hộp" (boxes) có chu kỳ ôn tập khác nhau (VD: Hộp 1 ôn mỗi ngày, Hộp 2 ôn mỗi 3 ngày). |
| **Streak** | Số ngày liên tiếp người dùng đạt được **Daily Goal**. Khuyến khích sự chăm chỉ. |
| **Daily Goal** | Mục tiêu học tập hàng ngày (VD: Học 10 từ mới, hoặc Ôn tập 50 từ). |
| **Saved Word** | Từ vựng mà người dùng chủ động đánh dấu/lưu lại vào danh sách cá nhân để học. |
| **AI Tutor** | Trợ lý ảo AI đóng vai trò như giáo viên, hỗ trợ giải nghĩa, sửa lỗi ngữ pháp và tạo bài tập cá nhân hoá. |
| **Contextual Learning** | Học từ vựng theo ngữ cảnh. Thay vì chỉ học nghĩa từ rời rạc, AI sẽ sinh ra các ví dụ gắn liền với câu nói hoặc tình huống cụ thể. |
| **Clean Architecture** | Kiến trúc mã nguồn chia thành các tầng (Layer) rõ rệt: Handler, Service, Domain, Repository để tối ưu khả năng bảo trì và test. |
| **Feature-based Architecture** | Kiến trúc mã nguồn Frontend gom nhóm các files theo tính năng (ví dụ `features/auth/`, `features/study/`) thay vì loại file. |