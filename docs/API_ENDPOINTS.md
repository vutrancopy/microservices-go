# 🌐 API ENDPOINTS - COMPREHENSIVE GUIDE

## 📍 Base URL
```
http://localhost:8080/v1
```

---

## 🔐 AUTHENTICATION

### Register
```http
POST /v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123",
  "userName": "username",
  "firstName": "John",
  "lastName": "Doe",
  "role": "rider"  // rider, driver, admin
}
```

### Login
```http
POST /v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123"
}

Response:
{
  "accessToken": "eyJhbGc...",
  "refreshToken": "eyJhbGc...",
  "user": { ... }
}
```

---

## 💰 PRICING (Fare Calculator)

### Get All Active Pricings
```http
GET /v1/pricing
```

### Get Pricing by Vehicle Type
```http
GET /v1/pricing/vehicle/:vehicleType

Example: GET /v1/pricing/vehicle/sedan
```

### Calculate Fare Estimate
```http
POST /v1/pricing/estimate
Content-Type: application/json

{
  "vehicleType": "sedan",
  "distanceKm": 15.5,
  "estimatedMinutes": 25,
  "waitingMinutes": 5,
  "surgeMultiplier": 1.5,
  "promoCode": "WELCOME20"
}

Response:
{
  "success": true,
  "data": {
    "baseFare": 3.00,
    "distanceFare": 23.25,
    "timeFare": 7.50,
    "waitingTimeFare": 2.50,
    "bookingFee": 2.00,
    "serviceFee": 5.74,
    "surgeCharge": 19.00,
    "discount": 10.00,
    "subtotal": 63.00,
    "total": 53.00,
    "currency": "USD"
  }
}
```

### Get Cancellation Fee
```http
POST /v1/pricing/cancellation-fee
Content-Type: application/json

{
  "vehicleType": "sedan",
  "minutesSinceBooking": 10
}

Response:
{
  "cancellationFee": 5.00,
  "vehicleType": "sedan"
}
```

### Update Pricing (Admin Only)
```http
PUT /v1/pricing/:id
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "baseFare": 3.50,
  "costPerKm": 1.75,
  "surgeMultiplier": 2.0
}
```

---

## 🎫 PROMO CODES

### Get All Active Promos
```http
GET /v1/promo/active

Response:
{
  "success": true,
  "data": [
    {
      "code": "WELCOME20",
      "description": "20% off your first ride",
      "type": "percentage",
      "value": 20.0,
      "maxDiscount": 10.0
    }
  ]
}
```

### Validate Promo Code
```http
POST /v1/promo/validate
Content-Type: application/json

{
  "code": "WELCOME20",
  "userId": 123,
  "rideFare": 50.00
}

Response:
{
  "success": true,
  "data": {
    "promoCode": { ... },
    "discount": 10.00,
    "originalFare": 50.00,
    "finalFare": 40.00
  }
}
```

### Apply Promo Code
```http
POST /v1/promo/apply
Content-Type: application/json

{
  "promoCodeId": 1,
  "userId": 123,
  "rideId": 456,
  "discountAmount": 10.00
}
```

### Get User Promo History
```http
GET /v1/promo/history/:userId

Example: GET /v1/promo/history/123
```

### Create Promo Code (Admin Only)
```http
POST /v1/promo
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "code": "SUMMER2024",
  "description": "Summer special discount",
  "type": "percentage",
  "value": 15.0,
  "maxDiscount": 20.0,
  "minRideAmount": 10.0,
  "maxUsagePerUser": 5,
  "maxTotalUsage": 1000,
  "validFrom": "2024-06-01T00:00:00Z",
  "validUntil": "2024-08-31T23:59:59Z",
  "firstRideOnly": false,
  "active": true
}
```

---

## 📍 FAVORITE LOCATIONS

### Create Favorite Location
```http
POST /v1/location/favorite
Authorization: Bearer {user_token}
Content-Type: application/json

{
  "label": "Home",
  "type": "home",
  "address": "123 Main St, San Francisco, CA",
  "latitude": 37.7749,
  "longitude": -122.4194,
  "isPrimary": true
}
```

### Get User Favorites
```http
GET /v1/location/favorite/user/:userId

Example: GET /v1/location/favorite/user/123
```

### Update Favorite Location
```http
PUT /v1/location/favorite/:id
Content-Type: application/json

{
  "label": "Mom's House",
  "isPrimary": true
}
```

### Delete Favorite Location
```http
DELETE /v1/location/favorite/:id
```

### Set As Primary
```http
PUT /v1/location/favorite/:id/primary
Authorization: Bearer {user_token}
```

---

## 📄 DRIVER DOCUMENTS

### Upload Document
```http
POST /v1/document
Authorization: Bearer {driver_token}
Content-Type: application/json

{
  "driverId": 456,
  "documentType": "license",
  "documentNumber": "DL123456789",
  "fileUrl": "https://s3.amazonaws.com/bucket/license.pdf",
  "expiryDate": "2026-12-31",
  "notes": "California driver's license"
}

Document Types:
- license
- insurance
- registration
- background_check
- profile_photo
- vehicle_photo
- identification
```

### Get Driver Documents
```http
GET /v1/document/driver/:driverId

Example: GET /v1/document/driver/456
```

### Get Document Summary
```http
GET /v1/document/driver/:driverId/summary

Response:
{
  "driverId": 456,
  "allDocumentsValid": true,
  "licenseVerified": true,
  "insuranceVerified": true,
  "registrationVerified": true,
  "backgroundVerified": true,
  "profilePhotoUploaded": true,
  "vehiclePhotoUploaded": true,
  "pendingCount": 0,
  "rejectedCount": 0
}
```

### Approve Document (Admin Only)
```http
POST /v1/document/:id/approve
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "verifiedBy": 1
}
```

### Reject Document (Admin Only)
```http
POST /v1/document/:id/reject
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "reason": "Document is blurry, please re-upload",
  "verifiedBy": 1
}
```

### Get Pending Documents (Admin Only)
```http
GET /v1/document/pending
Authorization: Bearer {admin_token}
```

---

## 📅 SCHEDULED RIDES

### Create Scheduled Ride
```http
POST /v1/scheduled
Authorization: Bearer {user_token}
Content-Type: application/json

{
  "vehicleType": "sedan",
  "pickupLatitude": 37.7749,
  "pickupLongitude": -122.4194,
  "pickupAddress": "123 Main St, San Francisco",
  "dropoffLatitude": 37.8044,
  "dropoffLongitude": -122.2712,
  "dropoffAddress": "456 Market St, Oakland",
  "scheduledTime": "2024-12-25T10:00:00Z",
  "estimatedDistance": 15.5,
  "estimatedDuration": 25,
  "estimatedFare": 45.00,
  "notes": "Please call when arriving",
  "promoCode": "WELCOME20"
}

Constraints:
- Must be 30 minutes to 7 days in advance
- Cannot schedule past 7 days
```

### Get User Scheduled Rides
```http
GET /v1/scheduled/user/:userId

Example: GET /v1/scheduled/user/123
```

### Get Upcoming Scheduled Rides
```http
GET /v1/scheduled/upcoming?limit=50
```

### Cancel Scheduled Ride
```http
POST /v1/scheduled/:id/cancel
Authorization: Bearer {user_token}
Content-Type: application/json

{
  "reason": "Plans changed"
}
```

### Update Scheduled Ride
```http
PUT /v1/scheduled/:id
Content-Type: application/json

{
  "notes": "Updated instructions",
  "promoCode": "SAVE5"
}
```

---

## 💼 DRIVER WALLET & EARNINGS

### Get Driver Wallet
```http
GET /v1/wallet/driver/:driverId

Response:
{
  "id": 1,
  "driverId": 456,
  "currentBalance": 1250.50,
  "totalEarnings": 5432.00,
  "totalWithdrawals": 4181.50,
  "pendingAmount": 0.00,
  "currency": "USD",
  "active": true
}
```

### Get Wallet Balance
```http
GET /v1/wallet/driver/:driverId/balance

Response:
{
  "driverId": 456,
  "balance": 1250.50
}
```

### Get Transaction History
```http
GET /v1/wallet/driver/:driverId/transactions?limit=50

Response:
{
  "success": true,
  "data": [
    {
      "id": 789,
      "type": "earning",
      "amount": 45.50,
      "balanceBefore": 1205.00,
      "balanceAfter": 1250.50,
      "description": "Ride earnings from ride #123",
      "processedAt": "2024-12-01T15:30:00Z"
    }
  ]
}
```

### Request Withdrawal
```http
POST /v1/wallet/driver/:driverId/withdraw
Authorization: Bearer {driver_token}
Content-Type: application/json

{
  "amount": 500.00,
  "method": "bank_transfer",
  "bankAccountInfo": "encrypted_account_details",
  "notes": "Weekly withdrawal"
}

Withdrawal Methods:
- bank_transfer
- paypal
- cash
- check

Constraints:
- Minimum: $10
- Must have sufficient balance
```

### Get Withdrawal Requests
```http
GET /v1/wallet/driver/:driverId/withdrawals

Response: List of withdrawal requests with status
```

### Get Pending Withdrawals (Admin Only)
```http
GET /v1/wallet/withdrawals/pending
Authorization: Bearer {admin_token}
```

### Approve Withdrawal (Admin Only)
```http
POST /v1/wallet/withdrawals/:id/approve
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "processedBy": 1
}
```

### Reject Withdrawal (Admin Only)
```http
POST /v1/wallet/withdrawals/:id/reject
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "reason": "Insufficient documentation",
  "processedBy": 1
}
```

---

## 🚗 VEHICLES

### Create Vehicle
```http
POST /v1/vehicle
Authorization: Bearer {driver_token}
Content-Type: application/json

{
  "driverId": 456,
  "vehicleType": "sedan",
  "make": "Toyota",
  "model": "Camry",
  "year": 2022,
  "color": "Black",
  "licensePlate": "ABC123",
  "capacity": 4
}
```

### Get All Vehicles
```http
GET /v1/vehicle?page=1&pageSize=10
```

### Get Vehicle by ID
```http
GET /v1/vehicle/:id
```

### Get Vehicles by Driver
```http
GET /v1/vehicle/driver/:driverId
```

### Activate/Deactivate Vehicle
```http
PUT /v1/vehicle/:id/activate
PUT /v1/vehicle/:id/deactivate
```

---

## 🚕 RIDES

### Create Ride
```http
POST /v1/ride
Authorization: Bearer {rider_token}
Content-Type: application/json

{
  "riderId": 123,
  "vehicleType": "sedan",
  "pickupLatitude": 37.7749,
  "pickupLongitude": -122.4194,
  "pickupAddress": "123 Main St",
  "dropoffLatitude": 37.8044,
  "dropoffLongitude": -122.2712,
  "dropoffAddress": "456 Market St",
  "promoCode": "WELCOME20"
}
```

### Match Driver Automatically
```http
POST /v1/ride/:id/match-driver
Authorization: Bearer {rider_token}

Algorithm considers:
- Driver proximity (< 5km)
- Vehicle availability & type match
- Driver rating
- Current load
```

### Start Ride
```http
POST /v1/ride/:id/start
Authorization: Bearer {driver_token}
```

### Complete Ride
```http
POST /v1/ride/:id/complete
Authorization: Bearer {driver_token}
Content-Type: application/json

{
  "actualDistance": 16.2,
  "actualDuration": 28,
  "waitingTime": 3
}
```

### Cancel Ride
```http
POST /v1/ride/:id/cancel
Content-Type: application/json

{
  "reason": "Rider cancelled",
  "cancelledBy": "rider"
}
```

### Rate Driver/Rider
```http
POST /v1/ride/:id/rate-driver
Content-Type: application/json

{
  "rating": 5,
  "comment": "Excellent service!"
}

POST /v1/ride/:id/rate-rider
```

---

## 💳 PAYMENTS

### Create Payment
```http
POST /v1/payment
Content-Type: application/json

{
  "rideId": 789,
  "riderId": 123,
  "driverId": 456,
  "amount": 45.50,
  "method": "credit_card"
}
```

### Get Payment by Ride
```http
GET /v1/payment/ride/:rideId
```

### Process Payment
```http
POST /v1/payment/:id/process
Authorization: Bearer {admin_token}
```

### Refund Payment
```http
POST /v1/payment/:id/refund
Authorization: Bearer {admin_token}
Content-Type: application/json

{
  "reason": "Service issue"
}
```

### Get Driver Earnings
```http
GET /v1/payment/driver/:driverId/earnings
```

---

## 👤 USERS

### Get User Profile
```http
GET /v1/user/:id
```

### Update User
```http
PUT /v1/user/:id
Content-Type: application/json

{
  "firstName": "John",
  "lastName": "Updated",
  "phoneNumber": "+1234567890"
}
```

### Update User Location (for drivers)
```http
PUT /v1/user/:id/location
Content-Type: application/json

{
  "latitude": 37.7749,
  "longitude": -122.4194
}
```

### Toggle Driver Availability
```http
PUT /v1/user/:id/toggle-availability
Authorization: Bearer {driver_token}
```

---

## 📊 SEARCH ENDPOINTS

### Search Rides
```http
POST /v1/ride/search
Content-Type: application/json

{
  "page": 1,
  "pageSize": 20,
  "filters": {
    "status": "completed",
    "riderId": 123
  }
}
```

### Search Vehicles
```http
POST /v1/vehicle/search
```

### Search Payments
```http
POST /v1/payment/search
```

---

## 🔧 ADMIN ENDPOINTS

### Get Pending Documents
```http
GET /v1/document/pending
Authorization: Bearer {admin_token}
```

### Get Pending Withdrawals
```http
GET /v1/wallet/withdrawals/pending
Authorization: Bearer {admin_token}
```

### Update Pricing Config
```http
PUT /v1/pricing/:id
Authorization: Bearer {admin_token}
```

### Create/Update Promo Codes
```http
POST /v1/promo
PUT /v1/promo/:id
DELETE /v1/promo/:id
Authorization: Bearer {admin_token}
```

---

## 📱 TYPICAL USER FLOWS

### 1. Rider Books a Ride with Promo
```
1. GET /v1/pricing/estimate (get fare estimate)
2. POST /v1/promo/validate (check promo code)
3. POST /v1/ride (create ride with promoCode)
4. POST /v1/ride/:id/match-driver (auto-match)
5. Wait for driver to start
6. Ride completes automatically
7. POST /v1/ride/:id/rate-driver
```

### 2. Driver Completes Ride & Withdraws Earnings
```
1. POST /v1/ride/:id/start
2. POST /v1/ride/:id/complete (earnings added to wallet)
3. GET /v1/wallet/driver/:id (check balance)
4. POST /v1/wallet/driver/:id/withdraw
5. Admin approves: POST /v1/wallet/withdrawals/:id/approve
6. Money transferred
```

### 3. Driver Verification Process
```
1. POST /v1/document (upload license)
2. POST /v1/document (upload insurance)
3. POST /v1/document (upload registration)
4. POST /v1/document (upload background check)
5. POST /v1/document (upload profile photo)
6. POST /v1/document (upload vehicle photo)
7. Admin reviews: GET /v1/document/pending
8. Admin approves: POST /v1/document/:id/approve
9. GET /v1/document/driver/:id/summary (check status)
```

### 4. Schedule Future Ride
```
1. POST /v1/scheduled (book for tomorrow 10 AM)
2. GET /v1/scheduled/user/:id (view upcoming)
3. System auto-matches driver 10 min before
4. Ride executes at scheduled time
5. Or cancel: POST /v1/scheduled/:id/cancel
```

---

## 🔑 AUTHENTICATION

Most endpoints require JWT authentication:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Roles:**
- `rider`: Can book rides, use promos, save locations
- `driver`: Can accept rides, upload documents, withdraw earnings
- `admin`: Full access to all endpoints

---

## ⚠️ ERROR RESPONSES

Standard error format:
```json
{
  "success": false,
  "error": {
    "type": "ValidationError",
    "message": "validation error",
    "details": "..."
  }
}
```

**Error Types:**
- `NotFound` (404)
- `ValidationError` (400)
- `NotAuthenticated` (401)
- `NotAuthorized` (403)
- `ResourceAlreadyExists` (409)
- `UnknownError` (500)

---

## 🎯 QUICK START

1. **Register as Rider:**
```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rider@test.com",
    "password": "Test@123",
    "userName": "test_rider",
    "firstName": "Test",
    "lastName": "Rider",
    "role": "rider"
  }'
```

2. **Login:**
```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "rider@test.com",
    "password": "Test@123"
  }'
```

3. **Get Active Promos:**
```bash
curl http://localhost:8080/v1/promo/active
```

4. **Calculate Fare:**
```bash
curl -X POST http://localhost:8080/v1/pricing/estimate \
  -H "Content-Type: application/json" \
  -d '{
    "vehicleType": "sedan",
    "distanceKm": 10,
    "estimatedMinutes": 20,
    "waitingMinutes": 2
  }'
```

---

## 📈 SUMMARY

**Total Endpoints:** ~80+ endpoints

**Categories:**
- Authentication: 3 endpoints
- Users: 7 endpoints
- Vehicles: 9 endpoints
- Rides: 16 endpoints
- Payments: 13 endpoints
- **Pricing: 5 endpoints** ✨
- **Promo: 7 endpoints** ✨
- **Locations: 6 endpoints** ✨
- **Documents: 8 endpoints** ✨
- **Scheduled: 6 endpoints** ✨
- **Wallet: 8 endpoints** ✨

**New Features Added:** 6 major modules  
**Total New Endpoints:** 40+ endpoints  

---

*Last Updated: 2025-11-13*  
*API Version: 2.0.0 - Enhanced Edition*
