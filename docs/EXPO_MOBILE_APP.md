# UNG Expo Mobile App (iOS/Android/Web)

> **Cross-platform mobile application using Expo and React Native**

## Overview

The Expo mobile app provides UNG functionality on iOS, Android, and Web platforms. Unlike the macOS app which uses the local SQLite database, the mobile app **connects to the Go API** for data access, enabling cloud sync and multi-device support.

## Core Architecture

### Technology Stack
- **Framework**: Expo SDK 50+
- **Language**: TypeScript
- **UI**: React Native with Expo Router
- **State Management**: Zustand or Redux Toolkit
- **API Client**: Axios or React Query
- **Database**: API-backed (no local SQLite)
- **Subscriptions**: RevenueCat
- **Authentication**: API JWT tokens

### Project Structure
```
ung-mobile/
├── app/                    # Expo Router screens
│   ├── (auth)/
│   │   ├── login.tsx
│   │   └── register.tsx
│   ├── (tabs)/
│   │   ├── dashboard.tsx
│   │   ├── clients.tsx
│   │   ├── invoices.tsx
│   │   ├── contracts.tsx
│   │   └── timer.tsx
│   └── _layout.tsx
├── components/             # Reusable components
│   ├── InvoiceCard.tsx
│   ├── ClientCard.tsx
│   ├── TimerWidget.tsx
│   └── ...
├── services/               # API and business logic
│   ├── api/
│   │   ├── client.ts
│   │   ├── auth.ts
│   │   ├── invoices.ts
│   │   └── ...
│   ├── storage/            # AsyncStorage helpers
│   └── subscriptions/      # RevenueCat integration
├── hooks/                  # Custom React hooks
├── store/                  # State management
├── types/                  # TypeScript types
├── utils/                  # Utility functions
└── constants/              # Constants and config
```

## Key Features

### 1. Authentication
```typescript
// services/api/auth.ts
interface AuthService {
  login(email: string, password: string): Promise<AuthResponse>
  register(email: string, password: string, name: string): Promise<AuthResponse>
  logout(): Promise<void>
  refreshToken(): Promise<string>
  getCurrentUser(): Promise<User>
}

interface AuthResponse {
  token: string
  refreshToken: string
  user: User
}
```

### 2. Dashboard
- Revenue summary (monthly/yearly)
- Quick actions (Create Invoice, Start Timer)
- Recent invoices
- Active contracts
- Revenue chart (Victory Native or similar)

### 3. Client Management
- Client list with search
- Client detail view
- Add/Edit client
- View client's contracts and invoices
- Quick actions (Create Invoice, New Contract)

### 4. Contract Management
- Active/Inactive lists
- Contract detail view
- Create/Edit contract
- Generate contract PDF (via API)
- Email contract

### 5. Invoice Management
- Invoice list with filters
- Invoice detail view
- Create invoice
- Generate PDF (via API)
- Email invoice
- Mark as paid/sent
- Photo attachments (upload to API)

### 6. Time Tracking
- Start/stop timer
- Timer running notification
- Background timer support
- Manual time entry
- Session list
- Convert to invoice

### 7. Settings
- Account settings
- Subscription management (RevenueCat)
- Language preferences
- Email template selection
- Logout

## API Integration

### API Client Setup
```typescript
// services/api/client.ts
import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';

const API_BASE_URL = 'https://api.ung.dev/v1';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: Add auth token
apiClient.interceptors.request.use(async (config) => {
  const token = await AsyncStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: Handle 401, refresh token
apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      // Try to refresh token
      const refreshed = await refreshAuthToken();
      if (refreshed) {
        // Retry original request
        return apiClient(error.config);
      } else {
        // Logout user
        await logout();
      }
    }
    return Promise.reject(error);
  }
);

export default apiClient;
```

### API Endpoints
```typescript
// services/api/invoices.ts
import apiClient from './client';
import { Invoice, InvoiceCreateInput } from '@/types';

export const invoicesAPI = {
  list: async (): Promise<Invoice[]> => {
    const { data } = await apiClient.get('/invoices');
    return data;
  },

  get: async (id: number): Promise<Invoice> => {
    const { data } = await apiClient.get(`/invoices/${id}`);
    return data;
  },

  create: async (input: InvoiceCreateInput): Promise<Invoice> => {
    const { data } = await apiClient.post('/invoices', input);
    return data;
  },

  update: async (id: number, input: Partial<Invoice>): Promise<Invoice> => {
    const { data } = await apiClient.put(`/invoices/${id}`, input);
    return data;
  },

  delete: async (id: number): Promise<void> => {
    await apiClient.delete(`/invoices/${id}`);
  },

  generatePDF: async (id: number): Promise<{ url: string }> => {
    const { data } = await apiClient.post(`/invoices/${id}/pdf`);
    return data;
  },

  email: async (id: number, to: string): Promise<void> => {
    await apiClient.post(`/invoices/${id}/email`, { to });
  },
};

// Similar for clients, contracts, tracking, etc.
```

## State Management

### Zustand Store Example
```typescript
// store/invoiceStore.ts
import { create } from 'zustand';
import { invoicesAPI } from '@/services/api/invoices';
import { Invoice } from '@/types';

interface InvoiceStore {
  invoices: Invoice[];
  loading: boolean;
  error: string | null;

  fetchInvoices: () => Promise<void>;
  createInvoice: (input: InvoiceCreateInput) => Promise<Invoice>;
  updateInvoice: (id: number, input: Partial<Invoice>) => Promise<void>;
  deleteInvoice: (id: number) => Promise<void>;
}

export const useInvoiceStore = create<InvoiceStore>((set, get) => ({
  invoices: [],
  loading: false,
  error: null,

  fetchInvoices: async () => {
    set({ loading: true, error: null });
    try {
      const invoices = await invoicesAPI.list();
      set({ invoices, loading: false });
    } catch (error) {
      set({ error: error.message, loading: false });
    }
  },

  createInvoice: async (input) => {
    const invoice = await invoicesAPI.create(input);
    set((state) => ({ invoices: [invoice, ...state.invoices] }));
    return invoice;
  },

  updateInvoice: async (id, input) => {
    const updated = await invoicesAPI.update(id, input);
    set((state) => ({
      invoices: state.invoices.map((inv) =>
        inv.id === id ? updated : inv
      ),
    }));
  },

  deleteInvoice: async (id) => {
    await invoicesAPI.delete(id);
    set((state) => ({
      invoices: state.invoices.filter((inv) => inv.id !== id),
    }));
  },
}));
```

## UI Components

### Invoice Card
```typescript
// components/InvoiceCard.tsx
import { View, Text, TouchableOpacity } from 'react-native';
import { Invoice } from '@/types';

interface Props {
  invoice: Invoice;
  onPress: () => void;
}

export function InvoiceCard({ invoice, onPress }: Props) {
  return (
    <TouchableOpacity onPress={onPress} className="bg-white p-4 rounded-lg shadow mb-2">
      <View className="flex-row justify-between items-center">
        <View>
          <Text className="text-lg font-semibold">{invoice.invoiceNum}</Text>
          <Text className="text-gray-600">{invoice.clientName}</Text>
        </View>
        <View className="items-end">
          <Text className="text-xl font-bold">
            {invoice.amount} {invoice.currency}
          </Text>
          <View className={`px-2 py-1 rounded ${statusColors[invoice.status]}`}>
            <Text className="text-xs font-medium">{invoice.status}</Text>
          </View>
        </View>
      </View>
    </TouchableOpacity>
  );
}
```

### Timer Widget
```typescript
// components/TimerWidget.tsx
import { useState, useEffect } from 'react';
import { View, Text, TouchableOpacity } from 'react-native';
import { useTimerStore } from '@/store/timerStore';

export function TimerWidget() {
  const { activeSession, startTimer, stopTimer } = useTimerStore();
  const [elapsed, setElapsed] = useState(0);

  useEffect(() => {
    if (!activeSession) return;

    const interval = setInterval(() => {
      const elapsed = Date.now() - activeSession.startTime.getTime();
      setElapsed(elapsed);
    }, 1000);

    return () => clearInterval(interval);
  }, [activeSession]);

  const formatTime = (ms: number) => {
    const seconds = Math.floor(ms / 1000);
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    return `${hours.toString().padStart(2, '0')}:${minutes
      .toString()
      .padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <View className="bg-blue-500 p-4 rounded-lg">
      {activeSession ? (
        <>
          <Text className="text-white text-sm">Active Timer</Text>
          <Text className="text-white text-3xl font-bold my-2">
            {formatTime(elapsed)}
          </Text>
          <Text className="text-white text-sm mb-3">{activeSession.project}</Text>
          <TouchableOpacity
            onPress={stopTimer}
            className="bg-white rounded py-2"
          >
            <Text className="text-blue-500 text-center font-semibold">
              Stop Timer
            </Text>
          </TouchableOpacity>
        </>
      ) : (
        <TouchableOpacity onPress={startTimer} className="py-2">
          <Text className="text-white text-center font-semibold">
            Start Timer
          </Text>
        </TouchableOpacity>
      )}
    </View>
  );
}
```

## Background Timer

### iOS Background Modes
```json
// app.json
{
  "expo": {
    "ios": {
      "infoPlist": {
        "UIBackgroundModes": ["location", "fetch"]
      }
    }
  }
}
```

### Timer Service
```typescript
// services/timer.ts
import * as TaskManager from 'expo-task-manager';
import * as BackgroundFetch from 'expo-background-fetch';

const TIMER_TASK = 'TIMER_BACKGROUND_TASK';

TaskManager.defineTask(TIMER_TASK, async () => {
  // Update timer in background
  const activeSession = await getActiveSession();
  if (activeSession) {
    await updateSession(activeSession.id, {
      duration: Date.now() - activeSession.startTime,
    });
  }
  return BackgroundFetch.BackgroundFetchResult.NewData;
});

export async function registerBackgroundTimer() {
  await BackgroundFetch.registerTaskAsync(TIMER_TASK, {
    minimumInterval: 60, // Update every minute
  });
}
```

## Push Notifications

### Expo Notifications Setup
```typescript
// services/notifications.ts
import * as Notifications from 'expo-notifications';
import * as Device from 'expo-device';

export async function registerForPushNotifications() {
  if (!Device.isDevice) return null;

  const { status: existingStatus } = await Notifications.getPermissionsAsync();
  let finalStatus = existingStatus;

  if (existingStatus !== 'granted') {
    const { status } = await Notifications.requestPermissionsAsync();
    finalStatus = status;
  }

  if (finalStatus !== 'granted') {
    return null;
  }

  const token = await Notifications.getExpoPushTokenAsync();

  // Send token to API
  await apiClient.post('/users/push-token', { token: token.data });

  return token.data;
}

// Handle notifications
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: true,
  }),
});
```

### Notification Types
- Timer reminders (running too long)
- Invoice due date approaching
- Payment received
- New client added
- Weekly revenue summary

## Revenue Cat Integration

### Setup
```typescript
// services/subscriptions/revenueCat.ts
import Purchases, {
  PurchasesOffering,
  CustomerInfo,
} from 'react-native-purchases';

const REVENUE_CAT_API_KEY_IOS = 'appl_...';
const REVENUE_CAT_API_KEY_ANDROID = 'goog_...';

export async function setupRevenueCat() {
  if (Platform.OS === 'ios') {
    Purchases.configure({ apiKey: REVENUE_CAT_API_KEY_IOS });
  } else if (Platform.OS === 'android') {
    Purchases.configure({ apiKey: REVENUE_CAT_API_KEY_ANDROID });
  }
}

export async function getOfferings(): Promise<PurchasesOffering[]> {
  const offerings = await Purchases.getOfferings();
  return offerings.all;
}

export async function purchasePackage(pkg: any): Promise<CustomerInfo> {
  const { customerInfo } = await Purchases.purchasePackage(pkg);
  return customerInfo;
}

export async function restorePurchases(): Promise<CustomerInfo> {
  const customerInfo = await Purchases.restorePurchases();
  return customerInfo;
}

export async function checkSubscriptionStatus(): Promise<boolean> {
  const customerInfo = await Purchases.getCustomerInfo();
  return customerInfo.entitlements.active['pro'] !== undefined;
}
```

### Subscription Screen
```typescript
// app/subscription.tsx
export default function SubscriptionScreen() {
  const [offerings, setOfferings] = useState<PurchasesOffering[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadOfferings();
  }, []);

  const loadOfferings = async () => {
    const offers = await getOfferings();
    setOfferings(offers);
  };

  const handlePurchase = async (pkg: any) => {
    setLoading(true);
    try {
      await purchasePackage(pkg);
      Alert.alert('Success', 'Subscription activated!');
    } catch (error) {
      Alert.alert('Error', error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <ScrollView className="flex-1 bg-white p-4">
      <Text className="text-2xl font-bold mb-4">Upgrade to Pro</Text>

      {offerings.map((offering) => (
        <View key={offering.identifier} className="mb-4">
          {offering.availablePackages.map((pkg) => (
            <TouchableOpacity
              key={pkg.identifier}
              onPress={() => handlePurchase(pkg)}
              className="bg-blue-500 p-4 rounded-lg mb-2"
            >
              <Text className="text-white font-bold text-lg">
                {pkg.product.title}
              </Text>
              <Text className="text-white">
                {pkg.product.priceString} / {pkg.packageType}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      ))}

      <TouchableOpacity onPress={restorePurchases} className="mt-4">
        <Text className="text-blue-500 text-center">Restore Purchases</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}
```

## Offline Support

### Offline Queue
```typescript
// services/offline.ts
import AsyncStorage from '@react-native-async-storage/async-storage';
import NetInfo from '@react-native-community/netinfo';

interface QueuedRequest {
  id: string;
  endpoint: string;
  method: string;
  data: any;
  timestamp: number;
}

export class OfflineQueue {
  private static QUEUE_KEY = 'offline_queue';

  static async enqueue(request: Omit<QueuedRequest, 'id' | 'timestamp'>) {
    const queue = await this.getQueue();
    const newRequest: QueuedRequest = {
      ...request,
      id: Date.now().toString(),
      timestamp: Date.now(),
    };
    queue.push(newRequest);
    await AsyncStorage.setItem(this.QUEUE_KEY, JSON.stringify(queue));
  }

  static async processQueue() {
    const isConnected = await NetInfo.fetch().then((state) => state.isConnected);
    if (!isConnected) return;

    const queue = await this.getQueue();
    for (const request of queue) {
      try {
        await apiClient.request({
          url: request.endpoint,
          method: request.method as any,
          data: request.data,
        });
        // Remove from queue on success
        await this.removeFromQueue(request.id);
      } catch (error) {
        console.error('Failed to process queued request:', error);
      }
    }
  }

  private static async getQueue(): Promise<QueuedRequest[]> {
    const json = await AsyncStorage.getItem(this.QUEUE_KEY);
    return json ? JSON.parse(json) : [];
  }

  private static async removeFromQueue(id: string) {
    const queue = await this.getQueue();
    const filtered = queue.filter((req) => req.id !== id);
    await AsyncStorage.setItem(this.QUEUE_KEY, JSON.stringify(filtered));
  }
}

// Listen for network changes
NetInfo.addEventListener((state) => {
  if (state.isConnected) {
    OfflineQueue.processQueue();
  }
});
```

## Web Support

### Expo Web Configuration
```typescript
// app.json
{
  "expo": {
    "web": {
      "favicon": "./assets/favicon.png",
      "bundler": "metro"
    }
  }
}
```

### Responsive Design
```typescript
// hooks/useResponsive.ts
import { useWindowDimensions } from 'react-native';

export function useResponsive() {
  const { width } = useWindowDimensions();

  return {
    isMobile: width < 768,
    isTablet: width >= 768 && width < 1024,
    isDesktop: width >= 1024,
  };
}

// Usage in components
function Dashboard() {
  const { isMobile } = useResponsive();

  return (
    <View className={isMobile ? 'flex-col' : 'flex-row'}>
      {/* Responsive layout */}
    </View>
  );
}
```

## Testing

### Unit Tests (Jest)
```typescript
// __tests__/services/api/invoices.test.ts
import { invoicesAPI } from '@/services/api/invoices';
import apiClient from '@/services/api/client';

jest.mock('@/services/api/client');

describe('invoicesAPI', () => {
  it('should list invoices', async () => {
    const mockInvoices = [{ id: 1, invoiceNum: 'INV-001' }];
    (apiClient.get as jest.Mock).mockResolvedValue({ data: mockInvoices });

    const invoices = await invoicesAPI.list();

    expect(invoices).toEqual(mockInvoices);
    expect(apiClient.get).toHaveBeenCalledWith('/invoices');
  });
});
```

### E2E Tests (Detox)
```typescript
// e2e/login.test.ts
describe('Login Flow', () => {
  beforeAll(async () => {
    await device.launchApp();
  });

  it('should login successfully', async () => {
    await element(by.id('email-input')).typeText('test@example.com');
    await element(by.id('password-input')).typeText('password123');
    await element(by.id('login-button')).tap();

    await expect(element(by.id('dashboard'))).toBeVisible();
  });
});
```

## Deployment

### EAS Build
```bash
# Install EAS CLI
npm install -g eas-cli

# Login
eas login

# Configure
eas build:configure

# Build for iOS
eas build --platform ios --profile production

# Build for Android
eas build --platform android --profile production

# Submit to stores
eas submit --platform ios
eas submit --platform android
```

### Environment Configuration
```typescript
// app.config.ts
export default {
  expo: {
    extra: {
      apiUrl: process.env.API_URL || 'https://api.ung.dev/v1',
      revenueCatApiKey: process.env.REVENUE_CAT_API_KEY,
    },
  },
};

// Usage in app
import Constants from 'expo-constants';

const API_URL = Constants.expoConfig?.extra?.apiUrl;
```

## Monetization

### Free Tier
- Basic invoice and client management
- Time tracking
- 10 invoices per month
- PDF generation

### Pro Tier ($9/month or $90/year)
- Unlimited invoices
- Recurring invoices
- Cloud sync
- Priority support
- Advanced reports
- Email delivery

### Business Tier ($29/month or $290/year)
- Everything in Pro
- Team collaboration
- API access
- White-label
- Dedicated support

## Key Differences from macOS App

1. **API-backed**: No local SQLite, uses Go API
2. **Cloud sync**: Data synced across devices
3. **Subscriptions**: RevenueCat for payments
4. **Push notifications**: Real-time updates
5. **Offline support**: Queue requests when offline
6. **Cross-platform**: iOS, Android, and Web

## Migration Checklist

- [ ] Expo project setup with TypeScript
- [ ] API client with auth interceptors
- [ ] All screens implemented
- [ ] State management (Zustand/Redux)
- [ ] RevenueCat integration
- [ ] Push notifications
- [ ] Offline queue
- [ ] Background timer
- [ ] E2E tests
- [ ] EAS build configuration
- [ ] App Store and Play Store listings

## Resources

- **Expo**: https://expo.dev
- **React Native**: https://reactnative.dev
- **RevenueCat**: https://www.revenuecat.com
- **React Navigation**: https://reactnavigation.org
- **Zustand**: https://github.com/pmndrs/zustand
- **NativeWind**: https://www.nativewind.dev (Tailwind for RN)
