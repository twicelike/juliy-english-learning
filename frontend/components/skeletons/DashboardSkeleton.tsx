export function DashboardSkeleton() {
  return (
    <div className="max-w-6xl mx-auto space-y-8 animate-pulse">
      {/* Profile Header Skeleton */}
      <div className="flex flex-col md:flex-row gap-6 bg-white/5 rounded-3xl p-6 md:p-8 border border-white/10">
        <div className="w-24 h-24 rounded-full bg-white/10 shrink-0"></div>
        <div className="flex-1 space-y-4">
          <div className="h-8 bg-white/10 rounded-md w-1/3"></div>
          <div className="h-4 bg-white/10 rounded-md w-1/4"></div>
          <div className="h-4 bg-white/10 rounded-md w-full max-w-md"></div>
        </div>
      </div>

      {/* Stats Grid Skeleton */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="bg-white/5 rounded-2xl p-4 border border-white/10 h-24 flex flex-col justify-between">
            <div className="h-4 bg-white/10 rounded w-1/2"></div>
            <div className="h-8 bg-white/10 rounded w-1/3 mt-2"></div>
          </div>
        ))}
      </div>

      {/* Sessions Skeleton */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="p-6 rounded-2xl bg-white/5 border border-white/10 h-32 flex flex-col gap-3">
            <div className="h-6 bg-white/10 rounded w-1/4"></div>
            <div className="h-4 bg-white/10 rounded w-3/4"></div>
            <div className="h-4 bg-white/10 rounded w-1/3 mt-auto"></div>
          </div>
        ))}
      </div>
    </div>
  );
}
