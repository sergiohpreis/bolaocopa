---
name: deploy
description: Deploy completo para produção (hostinger-vps). Assume que o código já está na origin/main. Faz git pull + docker compose up --build de todos os serviços.
---

SSH no servidor, faz pull da main e rebuilda todos os containers em produção.

Execute exatamente estes passos, sem pedir confirmação:

```
ssh hostinger-vps "cd /opt/bolaocopa && git pull origin main && docker compose --profile production up -d --build"
```

Reporte o resultado: quais containers foram recriados e se houve erro.
