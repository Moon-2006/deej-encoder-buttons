This fork adds mixed-input support for deej:

- **Encoders** send `ENC:<index>:<delta>` and are handled as pure PC-side deltas (no Arduino or deej absolute encoder state).
- **Buttons** send `BTN:<index>` and are mapped via `button_mapping` in config.
- **Potentiometers** keep the legacy `0-1023` pipe-delimited protocol (for direct absolute volume control).

For mixed setups, encoder indexes are offset by `NUM_POTENTIOMETERS` in the Arduino sketch so pots and encoders don't target the same deej slider index.

On encoder limit hits, deej sends `failed <index>` so the Arduino sketch can blink that encoder LED.
