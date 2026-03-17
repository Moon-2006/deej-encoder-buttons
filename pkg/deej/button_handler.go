package deej

import (
        "fmt"
        "strings"
        "strconv"
)

// handleButtonPress processa os comandos de botão recebidos do Arduino
func (d *Deej) handleButtonPress(buttonCommand string) {
        // Remove espaços em branco
        buttonCommand = strings.TrimSpace(buttonCommand)

        // Verifica se é um comando de botão (formato: BTN:0, BTN:1, etc)
        if !strings.HasPrefix(buttonCommand, "BTN:") {
                return
        }

        // Extrai o número do botão
        buttonNumStr := strings.TrimPrefix(buttonCommand, "BTN:")
        buttonNum := 0

        // Converte string para int
        if _, err := fmt.Sscanf(buttonNumStr, "%d", &buttonNum); err != nil {
                d.logger.Warnw("Failed to parse button number", "command", buttonCommand, "error", err)
                return
        }

        // DEBUG: Mostra o mapeamento completo
        d.logger.Debugw("Button mapping loaded", "mapping", d.config.ButtonMapping)

        // Busca a ação mapeada para este botão
        action, exists := d.config.ButtonMapping[buttonNum]
        if !exists {
                d.logger.Debugw("Button pressed but not mapped", "button", buttonNum)
                return
        }

        // Executa a ação
        d.executeButtonAction(action)
}

// executeButtonAction executa a ação configurada para o botão
func (d *Deej) executeButtonAction(action string) {
        d.logger.Debugw("Executing button action", "action", action)

        switch action {
        case "mute_master":
                d.toggleMuteMaster()
        case "media_play_pause":
                d.mediaPlayPause()
        case "media_next":
                d.mediaNext()
        case "media_previous":
                d.mediaPrevious()
        default:
                // Support custom key press actions using the key_ prefix.
                // For example: key_ctrl_shift_m will press Ctrl+Shift+M.
                if strings.HasPrefix(action, "key_") {
                        d.executeKeyAction(action)
                } else {
                        d.logger.Warnw("Unknown button action", "action", action)
                }
        }
}

// executeKeyAction parses a key_ action and sends the corresponding key presses.
// The format is key_<modifier>_<modifier>_<key>.  Supported modifiers are ctrl, shift and alt.
// Keys can be single letters (a–z) or function keys (f1–f24).  For example:
//
//   key_ctrl_shift_m  -> Ctrl+Shift+M
//   key_f13           -> F13
//
func (d *Deej) executeKeyAction(action string) {
        keysSpec := strings.TrimPrefix(action, "key_")
        parts := strings.Split(keysSpec, "_")
        codes := make([]uint16, 0, len(parts))
        for _, part := range parts {
                switch strings.ToLower(part) {
                case "ctrl":
                        codes = append(codes, uint16(0x11)) // VK_CONTROL
                case "shift":
                        codes = append(codes, uint16(0x10)) // VK_SHIFT
                case "alt":
                        codes = append(codes, uint16(0x12)) // VK_MENU
                default:
                        p := strings.ToLower(part)
                        if strings.HasPrefix(p, "f") && len(p) > 1 {
                                if n, err := strconv.Atoi(p[1:]); err == nil && n >= 1 && n <= 24 {
                                        codes = append(codes, uint16(0x70+n-1))
                                        continue
                                }
                        }
                        // Single letter keys
                        r := []rune(strings.ToUpper(part))
                        if len(r) == 1 && r[0] >= 'A' && r[0] <= 'Z' {
                                codes = append(codes, uint16(r[0]))
                        } else {
                                d.logger.Warnw("Unknown key spec part", "part", part)
                        }
                }
        }
        if len(codes) == 0 {
                d.logger.Warnw("No valid keys parsed for action", "action", action)
                return
        }
        if err := sendKeystrokes(codes); err != nil {
                d.logger.Warnw("Failed to send keystrokes", "keys", codes, "error", err)
        }
}