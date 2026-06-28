# Glossário — Bolão Copa

## Bolão
Um grupo de apostas criado por um Administrador, compartilhado via link de convite. Contém um conjunto de Participantes, as Regras de Pontuação e os Palpites de cada Participante.

## Administrador
O Participante que criou o Bolão. Tem poderes exclusivos: gerar o link de convite e configurar as Regras de Pontuação. Existe exatamente um Administrador por Bolão.

## Participante
Qualquer usuário autenticado que entrou no Bolão através do link de convite. A autenticação pode ser via Google OAuth ou Usuário com Senha. Pode registrar e alterar Palpites até o início de cada Jogo.

## Usuário com Senha
Método de autenticação onde o usuário cria uma conta diretamente no sistema com e-mail e senha, sem depender de um provedor externo como o Google OAuth.

## Jogo
Uma partida da Copa do Mundo, com horário de início, placar final e (quando definidos) dois times. Os dados de Jogos são sincronizados automaticamente de uma API externa. No Mata-mata, a API cria o Jogo antes do confronto ser decidido, com os times vazios; um Jogo nesse estado não aceita Palpite. O horário de início define o prazo para Palpites.

## Fase
A etapa do torneio à qual um Jogo pertence: Fase de Grupos, Oitavas, Quartas, Semifinal, Disputa de Terceiro ou Final. Vem crua da API externa (no código, campo `stage`) e o sistema apenas a exibe — não há lógica de chaveamento nem prazos, Pontuação ou Ranking dependentes de Fase. Serve para agrupar Jogos na exibição.
_Avoid_: stage, etapa, rodada

## Mata-mata
O conjunto das Fases eliminatórias (Oitavas, Quartas, Semifinal, Disputa de Terceiro, Final), em oposição à Fase de Grupos. Não é uma entidade do sistema — é o agrupamento que a Visão de Chaveamento representa.
_Avoid_: eliminatórias, knockout

## Visão de Chaveamento
Modo de exibição alternativo da aba Jogos (alternável por um toggle Lista/Chave) que mostra apenas os Jogos do Mata-mata, agrupados por Fase na ordem do torneio (16-avos → Oitavas → Quartas → Semifinal → Final/Disputa de Terceiro) e empilhados verticalmente — adequado ao layout mobile do app, sem árvore horizontal nem scroll lateral. Só fica disponível quando existe ao menos um Jogo de Mata-mata. Os confrontos de uma Fase ainda não decididos (Jogos sem times definidos) são resumidos numa única linha "N confrontos a definir" ao lado dos confrontos já conhecidos — sem Palpite.
_Avoid_: bracket, chaveamento, chave, árvore

## Palpite
O placar exato que um Participante prevê para um Jogo. Pode ser criado ou alterado até o momento em que o Jogo começa. Após o início do Jogo, o Palpite é bloqueado.

## Palpite Retroativo
Palpite registrado após o início de um Jogo, permitido apenas quando o Administrador habilita a funcionalidade no Bolão. Exige aprovação explícita do Administrador antes de ser computado no Ranking. Enquanto pendente, não é visível para os outros Participantes. Se aprovado, entra no cálculo de Pontuação normalmente; se rejeitado, é descartado.

## Resultado
O placar final de um Jogo, obtido da API externa. Usado para calcular a Pontuação de cada Palpite.

## Pontuação
Valor atribuído a um Palpite após o Resultado ser conhecido. Pode ser decimal (ex: 4.5).

**Fase de Grupos:**
- Placar exato: 10 pontos
- Acertou vencedor ou empate: 3 pontos
- Errou: 0 pontos

**Mata-mata:** a pontuação base da Fase de Grupos é multiplicada por um fator crescente por Fase.
- Acertou **quem avança** (time que o Participante chutou para vencer pelo placar): `3 × multiplicador`
- Acertou o **placar exato** do tempo normal: `10 × multiplicador` (e o time chutado avançou)
- Errou quem avança: 0 pontos

Multiplicadores: 16-avos 1.5×, Oitavas 2×, Quartas 2.5×, Semis 3×, Final 3.5×.

Jogos decididos nos pênaltis: o vencedor é determinado pelo campo `winner` da API externa — não pelo placar do tempo normal. Assim, um palpite de `2×0` para o Brasil conta como "acertou quem avança" se o Brasil passar, mesmo que o placar tenha sido `1×1` com pênaltis.

Palpite não registrado: 0 pontos (equivalente a errar).
_Avoid_: peso por fase, bônus de fase

## Ranking
Classificação global de todos os Participantes de um Bolão, ordenada pela soma de Pontuações. Único por Bolão — não há rankings por fase.

## Link de Convite
URL gerada pelo Administrador que permite a qualquer pessoa entrar no Bolão. Após acessar o link e autenticar via Google OAuth, o usuário se torna um Participante.

## Taxa de Entrada
Valor informativo que cada Participante deve pagar para participar do Bolão. Proposta pelo Administrador, definida por unanimidade dos Participantes presentes no momento da proposta. Imutável após definida. O sistema não processa pagamentos — apenas registra e exibe o valor acordado.

## Proposta de Taxa
Estado transitório criado pelo Administrador declarando um valor de Taxa de Entrada. Exige aprovação de todos os Participantes presentes no momento da proposta. Cancelada imediatamente se qualquer Participante votar não. Participantes que entrarem no Bolão após a criação da proposta não precisam votar.

## Jogo Ao Vivo
Um Jogo que já começou (`starts_at` no passado) mas ainda não teve o Resultado apurado (`finished = false`). Pode haver mais de um Jogo Ao Vivo simultaneamente (ex: última rodada da fase de grupos). Na aba Jogos do Bolão, Jogos Ao Vivo são exibidos em destaque fixo no topo, fora dos filtros de navegação.

## Jogo Encerrado
Um Jogo com `finished = true` — Resultado já apurado e Pontuações calculadas.

## Jogo Próximo
Um Jogo que ainda não começou (`starts_at` no futuro, `finished = false`). Inclui jogos com e sem Palpite registrado.

## Feed
Registro cronológico dos últimos 50 eventos relevantes de um Bolão, específico por Bolão. Atualizado via polling. Eventos incluem: palpite registrado (sem revelar o placar antes do Jogo começar), novo Participante entrou, Jogo começou (revela os palpites), Resultado apurado (revela pontuação de cada Participante). Não é persistido além dos 50 eventos mais recentes.

## Notificação
Mensagem enviada automaticamente pelo sistema ao Grupo Vinculado de um Bolão. Três tipos: **faltam dez minutos** (aviso para registrar Palpites antes do prazo), **partida iniciando** (apostas encerradas, boa sorte) e **fim de jogo** (placar final e quem pontuou). Disparadas pelo backend a cada ciclo de sincronização. O sistema garante no máximo um envio por tipo por Jogo — registrado na tabela `jogo_notifications` via `INSERT ... ON CONFLICT DO NOTHING`. O Administrador pode pausar ou retomar as Notificações a qualquer momento.

## Grupo Vinculado
Grupo do WhatsApp associado a um Bolão pelo Administrador. Destino de todas as Notificações daquele Bolão. Um Bolão tem no máximo um Grupo Vinculado. O vínculo é configurado pelo Administrador via painel de administração. Notificações de pré-jogo são enviadas para todos os Bolões que têm Grupo Vinculado configurado, independentemente de terem Palpites registrados no Jogo. O envio é feito por um serviço standalone (whatsmeow) chamado pelo backend a cada ciclo de sincronização (5 minutos); as janelas de disparo são [7min, 12min) antes do início para "faltam dez minutos" e [-2min, +2min) para "partida iniciando".
