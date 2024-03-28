# API Service for Guardian

### .env file configuration
```
SSH_HOST = ssh host ip
SSH_PORT = ssh port
SSH_USER = ssh username
SSH_KEYFILE = path to keyfile
DB_USER = database username
DB_PASS = database password
DB_HOST = database host name with port
DB_NAME = database name

API_PORT = port to run api on
API_KEY = api key to share with end user

```

## 1. GET - Get Usage By Account Number

**Description:** This request returns the following usage information:

1. Left Stove Cooking Time
2. Right Stove Cooking Time
3. Daily Cooking Time
4. Daily Power Consumption
5. Stove On/Off Count
6. Average Cooking Time Per Use
7. Average Power Consumption Per Use

Along with this information is the Calendar Date of the Data as well as the Unit Number for the Stove.

**Endpoint:** `example.com/:apikey/usage/:id`

---

## 2. GET - Get Stats By Account Number

**Description:** This request returns the following information:

1. Total Power Consumption
2. Average Daily Power Consumption

of the stove requested.

**Endpoint:** `example.com/:apikey/userstats/:id`

---

## 3. GET - Get All Stats

**Description:** This request returns the following information:

1. Unit Number
2. Total Power Consumption
3. Average Daily Power Consumption

for the entire fleet.

**Endpoint:** `example.com/:apikey/userstats/`
