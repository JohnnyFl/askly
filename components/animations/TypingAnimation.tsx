'use client';

import { useEffect, useState } from 'react';

interface TypingAnimationProps {
  texts: string[];
  typingSpeed?: number;
  deletingSpeed?: number;
  delayBetween?: number;
  className?: string;
}

export function TypingAnimation({
  texts,
  typingSpeed = 100,
  deletingSpeed = 50,
  delayBetween = 2000,
  className = ''
}: TypingAnimationProps) {
  const [displayedText, setDisplayedText] = useState('');
  const [currentIndex, setCurrentIndex] = useState(0);
  const [isDeleting, setIsDeleting] = useState(false);

  useEffect(() => {
    const currentText = texts[currentIndex];

    const timeout = setTimeout(() => {
      if (!isDeleting) {
        if (displayedText.length < currentText.length) {
          setDisplayedText(currentText.slice(0, displayedText.length + 1));
        } else {
          setTimeout(() => setIsDeleting(true), delayBetween);
        }
      } else {
        if (displayedText.length > 0) {
          setDisplayedText(currentText.slice(0, displayedText.length - 1));
        } else {
          setIsDeleting(false);
          setCurrentIndex((prev) => (prev + 1) % texts.length);
        }
      }
    }, isDeleting ? deletingSpeed : typingSpeed);

    return () => clearTimeout(timeout);
  }, [displayedText, isDeleting, currentIndex, texts, typingSpeed, deletingSpeed, delayBetween]);

  return (
    <span className={className}>
      {displayedText}
      <span className="animate-blink border-r-2 border-current ml-1"></span>
    </span>
  );
}
