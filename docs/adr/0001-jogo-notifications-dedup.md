# ADR 0001 — Deduplicação de notificações via banco

**Status:** aceito  
**Data:** 2026-06-20

## Contexto

O scheduler de notificações dispara a cada 1 minuto. As janelas de envio são de 4–5 minutos (`partida_iniciando`: [-2min, +2min); `faltam_dez_minutos`: [7min, 12min)), o que causava 4–5 mensagens idênticas por jogo no WhatsApp.

## Decisão

Usar uma tabela `jogo_notifications(jogo_id, notification_type)` com `PRIMARY KEY (jogo_id, notification_type)`. O envio é precedido por `INSERT ... ON CONFLICT DO NOTHING` — se retornar 0 rows, a notificação já foi enviada e é descartada.

## Alternativa rejeitada

Map em memória (`map[string]struct{}`): descartada porque não sobrevive a restarts do servidor. Se o processo reiniciar dentro da janela de 4 minutos, o map se perde e as notificações duplicam novamente.

## Consequências

- Schema de banco permanente — extensível para novos tipos de notificação sem mudança de estrutura.
- A deduplicação é por jogo, não por bolão: um `fim_de_jogo` inserido na tabela bloqueia o reenvio para todos os bolões, mesmo que novos bolões sejam configurados depois.
