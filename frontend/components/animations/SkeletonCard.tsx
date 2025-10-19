'use client';

export function SkeletonCard() {
  return (
    <div className="bg-white/90 backdrop-blur-md rounded-3xl shadow-xl border border-pink-100/50 p-6 animate-pulse">
      <div className="flex items-start gap-4 mb-4">
        <div className="h-20 w-20 bg-gradient-to-br from-gray-200 to-gray-300 rounded-full skeleton"></div>
        <div className="flex-1 space-y-3">
          <div className="h-6 bg-gradient-to-r from-gray-200 to-gray-300 rounded-xl w-3/4 skeleton"></div>
          <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg w-1/2 skeleton"></div>
        </div>
      </div>
      <div className="space-y-2 mb-4">
        <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg w-full skeleton"></div>
        <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg w-5/6 skeleton"></div>
        <div className="h-4 bg-gradient-to-r from-gray-200 to-gray-300 rounded-lg w-4/6 skeleton"></div>
      </div>
      <div className="h-12 bg-gradient-to-r from-gray-200 to-gray-300 rounded-xl w-full skeleton"></div>
    </div>
  );
}
