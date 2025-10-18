export const SUBSCRIPTION_TIERS = {
  FREE: {
    name: "Free",
    price: 0,
    priceId: null,
    questionsPerMonth: 10,
    features: [
      "10 questions per month",
      "Access to public channels",
      "Community FAQs",
      "Basic AI responses",
    ],
  },
  PRO: {
    name: "Pro",
    price: 9.99,
    priceId: process.env.NEXT_PUBLIC_POLAR_PRO_PRICE_ID,
    questionsPerMonth: 100,
    features: [
      "100 questions per month",
      "Access to all channels",
      "Priority AI responses",
      "Conversation history",
      "Advanced search",
    ],
  },
  PREMIUM: {
    name: "Premium",
    price: 19.99,
    priceId: process.env.NEXT_PUBLIC_POLAR_PREMIUM_PRICE_ID,
    questionsPerMonth: -1,
    features: [
      "Unlimited questions",
      "Access to exclusive channels",
      "Premium AI responses",
      "Full conversation history",
      "Advanced analytics",
      "Priority support",
    ],
  },
} as const;

export type SubscriptionTier = keyof typeof SUBSCRIPTION_TIERS;
