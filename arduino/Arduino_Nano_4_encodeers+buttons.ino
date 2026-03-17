// ========================================
// DEEJ WITH 4 ROTARY ENCODERS + BUTTONS (RELATIVE MODE)
// ========================================

/*
 * This sketch replaces the original analog‑style behaviour used by deej's
 * slider implementation. Instead of reading a fixed analog position from
 * each encoder and sending a "0|1023|..." formatted line, the encoders
 * now emit **relative** movement events. Each rotary encoder sends a
 * message of the form `ENC:<index>:<delta>` whenever it is turned. A
 * positive delta indicates a clockwise movement and a negative delta
 * indicates a counter‑clockwise movement. The amount of change is
 * represented by a single step (±1) but can be adjusted by changing
 * the `delta` assignment below. Buttons still emit `BTN:<index>` when
 * pressed.
 *
 * In addition, the sketch listens for responses from the host (deej).
 * When deej clamps a channel at its limit it sends back a line such as
 * `FAILED:2` to indicate that channel 2 hit a boundary. Upon receiving
 * this response, the Arduino briefly flashes an LED for the
 * corresponding encoder. By default all encoders share the built‑in LED
 * on pin 13; change the `LED_PINS` array if you have dedicated LEDs.
 */

const int NUM_ENCODERS = 4;

// Pin definitions for the four encoders (CLK/DT/SW for each)
#define ENC1_CLK 2
#define ENC1_DT  3
#define ENC1_SW  4

#define ENC2_CLK 5
#define ENC2_DT  6
#define ENC2_SW  7

#define ENC3_CLK 8
#define ENC3_DT  9
#define ENC3_SW  10

#define ENC4_CLK 11
#define ENC4_DT  12
#define ENC4_SW  A0

// Optional LED pins for each encoder.  If you have individual LEDs for
// each channel, specify their pins here.  By default all encoders share
// the built‑in LED on pin 13.
const int LED_PINS[NUM_ENCODERS] = {13, 13, 13, 13};

// Track the simulated analog position for each encoder.  This is used
// solely for clamping and is not transmitted directly.
int encPos[NUM_ENCODERS] = {512, 512, 512, 512};
int encLastCLK[NUM_ENCODERS];

// Button debouncing state
unsigned long lastButtonPress[NUM_ENCODERS] = {0, 0, 0, 0};
const unsigned long debounceDelay = 300;

void setup() {
  Serial.begin(9600);

  // Configure encoder pins
  pinMode(ENC1_CLK, INPUT_PULLUP);
  pinMode(ENC1_DT, INPUT_PULLUP);
  pinMode(ENC1_SW, INPUT_PULLUP);

  pinMode(ENC2_CLK, INPUT_PULLUP);
  pinMode(ENC2_DT, INPUT_PULLUP);
  pinMode(ENC2_SW, INPUT_PULLUP);

  pinMode(ENC3_CLK, INPUT_PULLUP);
  pinMode(ENC3_DT, INPUT_PULLUP);
  pinMode(ENC3_SW, INPUT_PULLUP);

  pinMode(ENC4_CLK, INPUT_PULLUP);
  pinMode(ENC4_DT, INPUT_PULLUP);
  pinMode(ENC4_SW, INPUT_PULLUP);

  // Configure LED pins
  for (int i = 0; i < NUM_ENCODERS; i++) {
    pinMode(LED_PINS[i], OUTPUT);
    digitalWrite(LED_PINS[i], LOW);
  }

  // Initialize last CLK states
  encLastCLK[0] = digitalRead(ENC1_CLK);
  encLastCLK[1] = digitalRead(ENC2_CLK);
  encLastCLK[2] = digitalRead(ENC3_CLK);
  encLastCLK[3] = digitalRead(ENC4_CLK);

  // Small startup delay to allow the host to open the serial port
  delay(1000);
}

// Main loop
void loop() {
  // Read encoders and emit relative deltas
  readEncoder(0, ENC1_CLK, ENC1_DT);
  readEncoder(1, ENC2_CLK, ENC2_DT);
  readEncoder(2, ENC3_CLK, ENC3_DT);
  readEncoder(3, ENC4_CLK, ENC4_DT);

  // Check buttons and emit BTN events
  checkButton(0, ENC1_SW);
  checkButton(1, ENC2_SW);
  checkButton(2, ENC3_SW);
  checkButton(3, ENC4_SW);

  // Handle any inbound responses from the host
  handleResponses();

  delay(1);
}

// Reads a rotary encoder and sends the delta as an `ENC` message when it moves.
void readEncoder(int index, int clkPin, int dtPin) {
  int currentCLK = digitalRead(clkPin);
  // Detect falling edge on CLK
  if (currentCLK != encLastCLK[index] && currentCLK == LOW) {
    int delta = 0;
    if (digitalRead(dtPin) != currentCLK) {
      // Clockwise
      delta = 1;
    } else {
      // Counter‑clockwise
      delta = -1;
    }

    // Update simulated analog position and clamp
    int newPos = encPos[index] + delta;
    if (newPos > 1023) newPos = 1023;
    if (newPos < 0)    newPos = 0;
    encPos[index] = newPos;

    // Emit relative movement to host
    Serial.print("ENC:");
    Serial.print(index);
    Serial.print(":");
    Serial.println(delta);
  }
  encLastCLK[index] = currentCLK;
}

// Debounced button check
void checkButton(int index, int swPin) {
  if (digitalRead(swPin) == LOW) {
    if (millis() - lastButtonPress[index] > debounceDelay) {
      Serial.print("BTN:");
      Serial.println(index);
      lastButtonPress[index] = millis();
    }
  }
}

// Blink the LED associated with an encoder.  If you have individual LEDs per
// encoder, set them in LED_PINS.  Otherwise the built‑in LED will blink.
void blinkLed(int index) {
  digitalWrite(LED_PINS[index], HIGH);
  delay(100);
  digitalWrite(LED_PINS[index], LOW);
}

// Read and handle inbound messages from the host.  Expected format:
// "FAILED:<index>" when a volume adjustment attempt hits the limit.
void handleResponses() {
  while (Serial.available() > 0) {
    String resp = Serial.readStringUntil('\n');
    resp.trim();
    // Check for failure notification (accept both "failed <idx>" and "FAILED:<idx>")
    if (resp.startsWith("failed") || resp.startsWith("FAILED")) {
      int idx = -1;
      // Determine delimiter position (space or colon)
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
        // Unknown index: blink all to indicate an error
        for (int i = 0; i < NUM_ENCODERS; i++) {
          blinkLed(i);
        }
      }
    }
  }
}