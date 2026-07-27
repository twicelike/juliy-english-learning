# Frontend Architecture (Next.js)

Dự án Frontend sử dụng **Next.js (App Router)** kết hợp với **Tailwind CSS**.

## Feature-based Architecture
Thay vì nhóm theo loại file (`components/`, `hooks/`, `api/`), chúng ta tổ chức mã nguồn theo **Features** (Chức năng). Điều này giúp code dễ scale, dễ tìm kiếm và độc lập.

```text
src/
├── app/                  # Next.js App Router (Pages, Layouts)
├── features/             # Feature-based modules
│   ├── auth/             # Chức năng Đăng nhập / Đăng ký
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── api/
│   │   └── types.ts
│   ├── study/            # Chức năng Học tập (Flashcards, Quizzes)
│   └── dictionary/       # Chức năng Tra từ / AI Explanation
├── components/           # Shared Components (Dùng chung toàn project)
│   ├── ui/               # UI Components (Buttons, Inputs, Modals - Radix/Shadcn)
│   └── layout/           # Navbar, Footer
├── lib/                  # Utils, cấu hình chung (Axios instance, Tailwind utils)
└── types/                # Global Types (User interface, API response format)
```

## State Management
- **Local/UI State:** Sử dụng React `useState`, `useReducer`.
- **Global State:** Sử dụng React Context (ví dụ: `AuthContext`, `ThemeContext`). Tránh lạm dụng Redux nếu không cần thiết.
- **Server State (Data Fetching):** Sử dụng **React Query (TanStack Query)** để cache data, quản lý loading/error state và tự động re-fetch.

## Thiết kế UI (Design System)
- Khuyến khích sử dụng **Tailwind CSS v4**.
- Áp dụng phong cách **Glassmorphism**, hiệu ứng mượt mà (Micro-animations).
- Các UI Components cơ bản (Nút bấm, Form) được đặt trong `src/components/ui` để tái sử dụng.