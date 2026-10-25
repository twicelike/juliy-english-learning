"use client";

import React, { useEffect, useState } from "react";
import apiClient from "@/lib/api-client";
import { Button } from "@/components/ui/Button";

interface User {
  id: string;
  email: string;
  display_name: string;
  role: string;
  status: string;
  created_at: string;
}

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchUsers = async () => {
    try {
      const res = await apiClient.get("/admin/users");
      if (res.data?.success) {
        setUsers(res.data.data?.users || []);
      }
    } catch (err) {
      console.error("Failed to fetch users", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleUpdateUser = async (id: string, role: string, status: string) => {
    try {
      const res = await apiClient.patch(`/admin/users/${id}`, {
        role,
        status: status || "active",
      });
      if (res.data?.success) {
        import("sonner").then(({ toast }) => toast.success("User updated successfully"));
        fetchUsers();
      }
    } catch (err: any) {
      console.error("Failed to update user", err);
      fetchUsers(); // Refresh to get true state
    }
  };

  if (loading) {
    return <div className="flex h-64 items-center justify-center">Loading users...</div>;
  }

  return (
    <div className="bg-white/5 border border-white/10 rounded-3xl p-8">
      <h2 className="text-xl font-semibold mb-6">User Management</h2>

      <div className="overflow-x-auto">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="border-b border-white/10">
              <th className="p-3 text-sm font-medium text-gray-400">Name</th>
              <th className="p-3 text-sm font-medium text-gray-400">Email</th>
              <th className="p-3 text-sm font-medium text-gray-400">Role</th>
              <th className="p-3 text-sm font-medium text-gray-400">Status</th>
              <th className="p-3 text-sm font-medium text-gray-400 text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map((user) => (
              <tr key={user.id} className="border-b border-white/5 hover:bg-white/5 transition-colors">
                <td className="p-3">{user.display_name}</td>
                <td className="p-3 text-gray-300">{user.email}</td>
                <td className="p-3">
                  <select
                    className="bg-black/50 border border-white/10 rounded px-2 py-1 outline-none text-sm focus:border-white/30 transition-colors"
                    value={user.role || "user"}
                    onChange={(e) => handleUpdateUser(user.id, e.target.value, user.status)}
                  >
                    <option value="user">User</option>
                    <option value="admin">Admin</option>
                    <option value="superior">Superior</option>
                  </select>
                </td>
                <td className="p-3">
                  <span
                    className={`px-2 py-1 text-xs rounded-full ${
                      user.status === "banned"
                        ? "bg-red-500/20 text-red-400"
                        : "bg-green-500/20 text-green-400"
                    }`}
                  >
                    {user.status === "banned" ? "Banned" : "Active"}
                  </span>
                </td>
                <td className="p-3 text-right">
                  <Button
                    size="sm"
                    variant={user.status === "banned" ? "secondary" : "primary"}
                    onClick={() =>
                      handleUpdateUser(
                        user.id,
                        user.role,
                        user.status === "banned" ? "active" : "banned"
                      )
                    }
                  >
                    {user.status === "banned" ? "Unban" : "Ban"}
                  </Button>
                </td>
              </tr>
            ))}
            {users.length === 0 && (
              <tr>
                <td colSpan={5} className="p-6 text-center text-gray-500">
                  No users found
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
