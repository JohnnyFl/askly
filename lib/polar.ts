export const SUBSCRIPTION_TIERS = {
  FREE: {
    name: "Free",
    price: 0,
    productId: null,
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
    productId: process.env.NEXT_PUBLIC_POLAR_PRO_PRODUCT_ID,
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
    productId: process.env.NEXT_PUBLIC_POLAR_PREMIUM_PRODUCT_ID,
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

export const POLAR_CONFIG = {
  organizationId: process.env.NEXT_PUBLIC_POLAR_ORGANIZATION_ID,
  checkoutUrl: 'https://polar.sh',
  customerPortalUrl: 'https://polar.sh/purchases',
};

export function createPolarCheckoutUrl(productId: string, customerEmail?: string): string {
  const baseUrl = `${POLAR_CONFIG.checkoutUrl}/${POLAR_CONFIG.organizationId}/products/${productId}/checkout`;

  if (customerEmail) {
    return `${baseUrl}?customer_email=${encodeURIComponent(customerEmail)}`;
  }

  return baseUrl;
}

export function getPolarCustomerPortalUrl(): string {
  return POLAR_CONFIG.customerPortalUrl;
}
