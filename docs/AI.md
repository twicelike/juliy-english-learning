# AI Strategy

Tích hợp OpenAI (GPT-4o / GPT-3.5) để tạo ra một "Gia sư Tiếng Anh ảo" thông minh.

## AI Use Cases (Tính năng)
1. **Word Explanation:** Giải thích nghĩa của từ dựa trên ngữ cảnh mà user bắt gặp từ đó.
2. **Sentence Correction:** Nhận diện và sửa lỗi ngữ pháp/từ vựng trong câu user tự viết.
3. **Grammar Feedback:** Đưa ra phản hồi chi tiết tại sao ngữ pháp sai và cách dùng đúng.
4. **Example Generation:** Sinh ra các ví dụ thực tế, đa dạng ngữ cảnh cho một từ vựng mới.
5. **Quiz Generation:** Tự động tạo câu hỏi trắc nghiệm (Multiple choice, Điền từ vào chỗ trống) để kiểm tra hiểu biết.
6. **Flashcard Generation:** Tự động trích xuất các từ khoá quan trọng từ một đoạn văn và tạo thành bộ Flashcards.
7. **Learning Recommendation:** AI phân tích `learning_progress` để gợi ý chủ đề học tiếp theo.
8. **Difficulty Adaptation:** AI tự động điều chỉnh độ khó của câu hỏi/ví dụ phù hợp với trình độ hiện tại của user.

## Prompt Strategy
- **System Prompt:** Phải định hình rõ vai trò (Role-playing) của AI.
  - *Ví dụ:* "You are a friendly, highly skilled English tutor for Vietnamese students. Explain English concepts simply in Vietnamese."
- **JSON Format Enforcement:** Trong mọi Prompt lấy dữ liệu về hệ thống (Quiz, Flashcard), phải yêu cầu Output chuẩn JSON để Backend Go có thể parse (ví dụ `json.Unmarshal`). Thêm cờ `response_format: { type: "json_object" }` khi gọi OpenAI API.

## Context Strategy
- Khi user gọi `Word Explanation`, hệ thống gửi kèm `level` của user để AI trả lời phù hợp (không dùng từ quá khó với người mới).
- Sử dụng collection `ai_chat` để gửi kèm lịch sử hội thoại (3-5 tin nhắn gần nhất) giúp AI giữ được ngữ cảnh đang thảo luận.