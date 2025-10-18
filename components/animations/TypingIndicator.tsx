'use client';

export function TypingIndicator() {
  return (
    <div className="flex items-center space-x-2 p-4 bg-gradient-to-br from-purple-50 to-pink-50 rounded-2xl border-2 border-purple-200/50 w-fit">
      <div className="flex space-x-1">
        <div className="w-2 h-2 bg-[#6D28D9] rounded-full animate-bounce"></div>
        <div className="w-2 h-2 bg-[#6D28D9] rounded-full animate-bounce animation-delay-200"></div>
        <div className="w-2 h-2 bg-[#6D28D9] rounded-full animate-bounce animation-delay-400"></div>
      </div>
      <span className="text-sm text-slate-600">AI is thinking...</span>
    </div>
  );
}
