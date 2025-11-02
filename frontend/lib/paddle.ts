// Subscription tier definitions (from SUBSCRIPTION_STRATEGY.md)
export const SUBSCRIPTION_TIERS = {
  FREE: {
    name: "Free",
    price: 0,
    priceIdMonthly: null,
    priceIdAnnual: null,
    questionsPerMonth: 5,
    channelsLimit: 1,
    features: [
      "5 questions per month",
      "Access to 1 channel (your choice)",
      "Excited AI responses",
      "Session-only history",
    ],
  },
  PRO: {
    name: "Pro",
    price: 7,
    annualPrice: 66, // $5.50/mo when paid annually
    priceIdMonthly: process.env.NEXT_PUBLIC_PADDLE_PRO_MONTHLY_PRICE_ID!,
    priceIdAnnual: process.env.NEXT_PUBLIC_PADDLE_PRO_ANNUAL_PRICE_ID!,
    questionsPerMonth: 50,
    channelsLimit: 2,
    features: [
      "50 questions per month (~2/day)",
      "Access to BOTH channels",
      "90-day conversation history",
      "Export conversations (markdown/PDF)",
      "Priority response speed",
    ],
  },
  PREMIUM: {
    name: "Premium",
    price: 15,
    annualPrice: 144, // $12/mo when paid annually
    priceIdMonthly: process.env.NEXT_PUBLIC_PADDLE_PREMIUM_MONTHLY_PRICE_ID!,
    priceIdAnnual: process.env.NEXT_PUBLIC_PADDLE_PREMIUM_ANNUAL_PRICE_ID!,
    questionsPerMonth: 200,
    channelsLimit: -1, // Unlimited
    features: [
      "200 questions per month (~7/day)",
      "Access to all channels + early access",
      "Unlimited conversation history + search",
      "Priority support (24h email response)",
      "Early access to new channels",
      "Advanced features when launched",
    ],
  },
} as const;

export type SubscriptionTier = keyof typeof SUBSCRIPTION_TIERS;

export const PADDLE_CONFIG = {
  vendorId: process.env.NEXT_PUBLIC_PADDLE_VENDOR_ID!,
  environment: (process.env.NEXT_PUBLIC_PADDLE_ENVIRONMENT || 'sandbox') as 'sandbox' | 'production',
  clientToken: process.env.NEXT_PUBLIC_PADDLE_CLIENT_TOKEN,
};

// Initialize Paddle.js
export function initializePaddle(): Promise<any> {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined') {
      resolve(null);
      return;
    }

    // Check if already loaded
    // @ts-ignore
    if (window.Paddle) {
      // @ts-ignore
      resolve(window.Paddle);
      return;
    }

    // Load Paddle.js script
    const script = document.createElement('script');
    script.src = 'https://cdn.paddle.com/paddle/v2/paddle.js';
    script.async = true;

    script.onload = () => {
      // @ts-ignore
      if (window.Paddle) {
        // @ts-ignore
        window.Paddle.Environment.set(PADDLE_CONFIG.environment);
        // @ts-ignore
        window.Paddle.Setup({
          vendor: parseInt(PADDLE_CONFIG.vendorId),
          eventCallback: function(data: any) {
            console.log('Paddle event:', data);
          }
        });
        // @ts-ignore
        resolve(window.Paddle);
      } else {
        reject(new Error('Paddle failed to load'));
      }
    };

    script.onerror = () => {
      reject(new Error('Failed to load Paddle.js'));
    };

    document.head.appendChild(script);
  });
}

// Open Paddle checkout overlay
export async function openPaddleCheckout(
  priceId: string,
  userEmail?: string,
  userId?: string
): Promise<void> {
  const paddle = await initializePaddle();

  if (!paddle) {
    console.error('Paddle not initialized');
    throw new Error('Paddle not initialized');
  }

  // @ts-ignore
  paddle.Checkout.open({
    items: [{ priceId, quantity: 1 }],
    customer: userEmail ? { email: userEmail } : undefined,
    customData: userId ? { user_id: userId } : undefined,
    successCallback: (data: any) => {
      console.log('Checkout successful:', data);
      // Redirect to account page with success message
      window.location.href = '/account?checkout=success';
    },
    closeCallback: (data: any) => {
      console.log('Checkout closed:', data);
    },
  });
}

// Create checkout via backend API (alternative approach)
export async function createCheckoutSession(
  priceId: string,
  email?: string
): Promise<string> {
  const token = localStorage.getItem('auth_token');

  const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/checkout`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
    },
    body: JSON.stringify({
      price_id: priceId,
      email,
    }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.error || 'Failed to create checkout');
  }

  const data = await response.json();
  return data.checkout_url;
}

// Get Paddle billing portal URL
export function getPaddleBillingPortalUrl(): string {
  return PADDLE_CONFIG.environment === 'sandbox'
    ? 'https://sandbox-subscription-management.paddle.com'
    : 'https://subscription-management.paddle.com';
}

// Helper: Get monthly price for display
export function getMonthlyPrice(tier: SubscriptionTier, billingCycle: 'monthly' | 'annual'): number {
  const tierData = SUBSCRIPTION_TIERS[tier];
  if (tier === 'FREE') return 0;

  if (billingCycle === 'annual') {
    const annualPrice = tier === 'PRO' ? SUBSCRIPTION_TIERS.PRO.annualPrice : SUBSCRIPTION_TIERS.PREMIUM.annualPrice;
    return annualPrice! / 12;
  }

  return tierData.price;
}

// Helper: Get annual savings
export function getAnnualSavings(tier: SubscriptionTier): number {
  if (tier === 'FREE') return 0;

  const tierData = SUBSCRIPTION_TIERS[tier];
  const annualPrice = tier === 'PRO' ? SUBSCRIPTION_TIERS.PRO.annualPrice : SUBSCRIPTION_TIERS.PREMIUM.annualPrice;

  if (!annualPrice) return 0;

  const monthlyTotal = tierData.price * 12;
  return monthlyTotal - annualPrice;
}
