// ========================================
// DEEJ WITH ROTARY ENCODERS + BUTTONS (RELATIVE MODE)
// ========================================

/*
 * This sketch sends relative encoder movement events to Deej in the format:
 *   ENC:<index>:<delta>
 *
 * and button presses in the format:
 *   BTN:<index>
 *
 * It also listens for failure messages from Deej like:
 *   failed 2
 * or
 *   FAILED:2
 *
 * and blinks the LED for that encoder.
 *
 * =========================
 * TO CHANGE ENCODER COUNT:
 * =========================
 * Only update:
 *   1) NUM_ENCODERS
 *   2) CLK_PINS
 *   3) DT_PINS
 *   4) SW_PINS
 *   5) LED_PINS
 */

// ===== USER CONFIG =====
const int NUM_ENCODERS = 1;

const int CLK_PINS[NUM_ENCODERS] = {2};
const int DT_PINS[NUM_ENCODERS]  = {3};
const int SW_PINS[NUM_ENCODERS]  = {4};

// Use one LED per encoder, or repeat 13 to share the built-in LED
const int LED_PINS[NUM_ENCODERS] = {13};
// =======================

// Internal state
int encPos[NUM_ENCODERS];
int encLastCLK[NUM_ENCODERS];
unsigned long lastButtonPress[NUM_ENCODERS];

const unsigned long debounceDelay = 300;

void setup() {
  Serial.begin(9600);

  for (int i = 0; i < NUM_ENCODERS; i++) {
    pinMode(CLK_PINS[i], INPUT_PULLUP);
    pinMode(DT_PINS[i], INPUT_PULLUP);
    pinMode(SW_PINS[i], INPUT_PULLUP);

    pinMode(LED_PINS[i], OUTPUT);
    digitalWrite(LED_PINS[i], LOW);

    encPos[i] = 512;
    encLastCLK[i] = digitalRead(CLK_PINS[i]);
    lastButtonPress[i] = 0;
  }

  // Give the PC time to open the serial connection
  delay(1000);
}

void loop() {
  for (int i = 0; i < NUM_ENCODERS; i++) {
    readEncoder(i);
    checkButton(i);
  }

  handleResponses();
  delay(1);
}

// Reads one encoder and sends a relative delta when it moves
void readEncoder(int index) {
  int clkPin = CLK_PINS[index];
  int dtPin  = DT_PINS[index];

  int currentCLK = digitalRead(clkPin);

  // Detect falling edge
  if (currentCLK != encLastCLK[index] && currentCLK == LOW) {
    int delta = (digitalRead(dtPin) != currentCLK) ? 5 : -5;

    // Track internal position only for clamping/reference
    int newPos = encPos[index] + delta;
    if (newPos > 1023) newPos = 1023;
    if (newPos < 0)    newPos = 0;
    encPos[index] = newPos;

    // Send encoder movement
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

// Blink LED for one encoder
void blinkLed(int index) {
  digitalWrite(LED_PINS[index], HIGH);
  delay(100);
  digitalWrite(LED_PINS[index], LOW);
}

// Read responses from Deej and blink the corresponding LED on failure
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
        // Unknown index: blink all LEDs
        for (int i = 0; i < NUM_ENCODERS; i++) {
          blinkLed(i);
        }
      }
    }
  }
}
