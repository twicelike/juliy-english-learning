import { ProgressQuality } from "@/lib/learning-api";

interface RatingBarProps {
  onRating: (rating: ProgressQuality) => void;
}

export function RatingBar({ onRating }: RatingBarProps) {
  return (
    <div className="w-full max-w-md mt-8">
      <div className="flex gap-3">
        <button
          onClick={() => onRating(1)}
          className="flex-1 py-3.5 rounded-2xl bg-red-500/10 border border-red-500/20 hover:bg-red-500/20 hover:border-red-500/30 transition-all group"
        >
          <div className="text-red-400 font-semibold group-hover:scale-105 transition-transform">Again</div>
          <div className="text-red-400/50 text-[10px] uppercase tracking-widest mt-1 font-mono">1</div>
        </button>
        <button
          onClick={() => onRating(3)}
          className="flex-1 py-3.5 rounded-2xl bg-orange-500/10 border border-orange-500/20 hover:bg-orange-500/20 hover:border-orange-500/30 transition-all group"
        >
          <div className="text-orange-400 font-semibold group-hover:scale-105 transition-transform">Hard</div>
          <div className="text-orange-400/50 text-[10px] uppercase tracking-widest mt-1 font-mono">3</div>
        </button>
        <button
          onClick={() => onRating(4)}
          className="flex-1 py-3.5 rounded-2xl bg-blue-500/10 border border-blue-500/20 hover:bg-blue-500/20 hover:border-blue-500/30 transition-all group"
        >
          <div className="text-blue-400 font-semibold group-hover:scale-105 transition-transform">Good</div>
          <div className="text-blue-400/50 text-[10px] uppercase tracking-widest mt-1 font-mono">4</div>
        </button>
        <button
          onClick={() => onRating(5)}
          className="flex-1 py-3.5 rounded-2xl bg-emerald-500/10 border border-emerald-500/20 hover:bg-emerald-500/20 hover:border-emerald-500/30 transition-all group"
        >
          <div className="text-emerald-400 font-semibold group-hover:scale-105 transition-transform">Easy</div>
          <div className="text-emerald-400/50 text-[10px] uppercase tracking-widest mt-1 font-mono">5</div>
        </button>
      </div>
    </div>
  );
}
