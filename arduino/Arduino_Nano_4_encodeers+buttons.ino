// ========================================
// DEEJ WITH MIXED INPUTS:
// - Rotary encoders (relative deltas)
// - Encoder buttons
// - Potentiometers (absolute 0-1023)
// ========================================

/*
 * Protocol:
 * - Encoder movement: ENC:<index>:<delta>
 * - Encoder button:   BTN:<index>
 * - Potentiometers:   standard deej line, e.g. "123|456|789"
 *
 * Failure response from deej:
 * - failed <index>
 * - FAILED:<index>
 *
 * On failure, this sketch blinks the LED for that encoder index.
 */

// ===== USER CONFIG =====
const int NUM_ENCODERS = 1;
const int NUM_POTENTIOMETERS = 0;

const int CLK_PINS[NUM_ENCODERS] = {2};
const int DT_PINS[NUM_ENCODERS]  = {3};
const int SW_PINS[NUM_ENCODERS]  = {4};
const int LED_PINS[NUM_ENCODERS] = {13};

const int POT_PINS[NUM_POTENTIOMETERS] = {};
// =======================

int encLastCLK[NUM_ENCODERS];
unsigned long lastButtonPress[NUM_ENCODERS];
int potValues[NUM_POTENTIOMETERS];

const unsigned long debounceDelay = 300;
const int ENCODER_STEP = 5;

void setup() {
  Serial.begin(9600);

  for (int i = 0; i < NUM_ENCODERS; i++) {
    pinMode(CLK_PINS[i], INPUT_PULLUP);
    pinMode(DT_PINS[i], INPUT_PULLUP);
    pinMode(SW_PINS[i], INPUT_PULLUP);

    pinMode(LED_PINS[i], OUTPUT);
    digitalWrite(LED_PINS[i], LOW);

    encLastCLK[i] = digitalRead(CLK_PINS[i]);
    lastButtonPress[i] = 0;
  }

  for (int i = 0; i < NUM_POTENTIOMETERS; i++) {
    pinMode(POT_PINS[i], INPUT);
    potValues[i] = 0;
  }

  // Give the PC time to open the serial connection
  delay(1000);
}

void loop() {
  for (int i = 0; i < NUM_ENCODERS; i++) {
    readEncoder(i);
    checkButton(i);
  }

  if (NUM_POTENTIOMETERS > 0) {
    readPotentiometers();
    sendPotentiometerValues();
  }

  handleResponses();
  delay(1);
}

// Reads one encoder and sends a relative delta when it moves
void readEncoder(int index) {
  int clkPin = CLK_PINS[index];
  int dtPin = DT_PINS[index];

  int currentCLK = digitalRead(clkPin);

  // Detect falling edge and emit relative movement only
  if (currentCLK != encLastCLK[index] && currentCLK == LOW) {
    int delta = (digitalRead(dtPin) != currentCLK) ? ENCODER_STEP : -ENCODER_STEP;

    Serial.print("ENC:");
    Serial.print(index);
    Serial.print(":");
    Serial.println(delta);
  }

  encLastCLK[index] = currentCLK;
}

// Checks one encoder button with debounce
void checkButton(int index) {
  int swPin = SW_PINS[index];

  if (digitalRead(swPin) == LOW) {
    if (millis() - lastButtonPress[index] > debounceDelay) {
      Serial.print("BTN:");
      Serial.println(index);
      lastButtonPress[index] = millis();
    }
  }
}

void readPotentiometers() {
  for (int i = 0; i < NUM_POTENTIOMETERS; i++) {
    potValues[i] = analogRead(POT_PINS[i]);
  }
}

void sendPotentiometerValues() {
  String builtString = "";

  for (int i = 0; i < NUM_POTENTIOMETERS; i++) {
    builtString += String(potValues[i]);

    if (i < NUM_POTENTIOMETERS - 1) {
      builtString += "|";
    }
  }

  Serial.println(builtString);
}

// Blink LED for one encoder
void blinkLed(int index) {
  digitalWrite(LED_PINS[index], HIGH);
  delay(100);
  digitalWrite(LED_PINS[index], LOW);
}

// Read responses from deej and blink the corresponding LED on failure
void handleResponses() {
  while (Serial.available() > 0) {
    String resp = Serial.readStringUntil('\n');
    resp.trim();

    if (resp.startsWith("failed") || resp.startsWith("FAILED")) {
      int idx = -1;

      int delimPos = resp.indexOf(' ');
      if (delimPos == -1) {
        delimPos = resp.indexOf(':');
      }

      if (delimPos != -1) {
        String idxStr = resp.substring(delimPos + 1);
        idx = idxStr.toInt();
      }

      if (idx >= 0 && idx < NUM_ENCODERS) {
        blinkLed(idx);
      } else {
        for (int i = 0; i < NUM_ENCODERS; i++) {
          blinkLed(i);
        }
      }
    }
  }
}
