# Database Design (MongoDB)

Dự án sử dụng MongoDB. Do tính chất Document Database của MongoDB, chúng ta ưu tiên sử dụng cấu trúc nhúng (Embedded Documents) cho những dữ liệu ít thay đổi hoặc có quan hệ chặt chẽ 1-N (One-to-Few), và sử dụng Tham chiếu (References) cho quan hệ N-N (Many-to-Many).

## Các Collections Chính

1. **`users`**: Lưu trữ thông tin tài khoản người dùng.
   - `_id`: ObjectID
   - `email`: String (Unique Index)
   - `password_hash`: String
   - `name`: String
   - `created_at`, `updated_at`: Timestamp

2. **`dictionary`**: Lưu trữ từ điển chung/từ điển hệ thống (nếu có) hoặc cached lại các từ đã được AI giải nghĩa để tiết kiệm chi phí gọi API.
   - `_id`: ObjectID
   - `word`: String (Unique Index)
   - `pos`: String (Part of speech)
   - `meanings`: Array of Objects
   - `phonetics`: String

3. **`saved_words`**: Danh sách từ vựng cá nhân mà mỗi user lưu lại.
   - `_id`: ObjectID
   - `user_id`: ObjectID (Reference to `users` - Index)
   - `word_id`: ObjectID (Reference to `dictionary` hoặc lưu trực tiếp String tuỳ chiến lược)
   - `added_at`: Timestamp

4. **`learning_progress`**: Lưu trữ tiến độ học của mỗi user đối với từng từ vựng (phục vụ Spaced Repetition).
   - `_id`: ObjectID
   - `user_id`: ObjectID (Reference to `users`)
   - `word`: String
   - `box`: Integer (Leitner system box 1-5)
   - `next_review_date`: Timestamp (Index để query các từ cần học hôm nay)

5. **`review_history`**: Nhật ký mỗi lần user ôn tập từ vựng (dùng để vẽ biểu đồ thống kê).
   - `_id`: ObjectID
   - `user_id`: ObjectID (Reference to `users`)
   - `word`: String
   - `result`: Boolean (Pass/Fail)
   - `reviewed_at`: Timestamp

6. **`daily_goal`**: Mục tiêu và streak học tập hàng ngày của user.
   - `_id`: ObjectID
   - `user_id`: ObjectID (Reference to `users`)
   - `date`: String (Format YYYY-MM-DD)
   - `words_learned`: Integer
   - `goal_target`: Integer

7. **`ai_chat`**: Lịch sử chat giữa User và AI Tutor để giữ ngữ cảnh (Context) cho các câu hỏi tiếp theo.
   - `_id`: ObjectID
   - `user_id`: ObjectID (Reference to `users`)
   - `messages`: Array of Objects `[{role: "user", content: "..."}, {role: "assistant", content: "..."}]`
   - `updated_at`: Timestamp

## Indexing Strategy
- `users`: `{ email: 1 }` (Unique)
- `learning_progress`: `{ user_id: 1, next_review_date: 1 }` (Compound Index để lấy nhanh danh sách từ cần học hôm nay của một user).
- `saved_words`: `{ user_id: 1, word_id: 1 }` (Unique Compound Index tránh lưu trùng từ).