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
Uma partida da Copa do Mundo, com dois times, horário de início e placar final. Os dados de Jogos são sincronizados automaticamente de uma API externa. O horário de início define o prazo para Palpites.

## Palpite
O placar exato que um Participante prevê para um Jogo. Pode ser criado ou alterado até o momento em que o Jogo começa. Após o início do Jogo, o Palpite é bloqueado.

## Palpite Retroativo
Palpite registrado após o início de um Jogo, permitido apenas quando o Administrador habilita a funcionalidade no Bolão. Exige aprovação explícita do Administrador antes de ser computado no Ranking. Enquanto pendente, não é visível para os outros Participantes. Se aprovado, entra no cálculo de Pontuação normalmente; se rejeitado, é descartado.

## Resultado
O placar final de um Jogo, obtido da API externa. Usado para calcular a Pontuação de cada Palpite.

## Pontuação
Valor atribuído a um Palpite após o Resultado ser conhecido:
- **Placar exato**: 10 pontos
- **Acertou vencedor ou empate**: 3 pontos
- **Errou**: 0 pontos
- **Palpite não registrado**: 0 pontos (equivalente a errar)

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
Mensagem enviada automaticamente pelo sistema ao Grupo Vinculado de um Bolão. Três tipos: **faltam dez minutos** (aviso para registrar Palpites antes do prazo), **partida iniciando** (apostas encerradas, boa sorte) e **fim de jogo** (placar final e quem pontuou). Disparadas pelo backend a cada ciclo de sincronização. O Administrador pode pausar ou retomar as Notificações a qualquer momento.

## Grupo Vinculado
Grupo do WhatsApp associado a um Bolão pelo Administrador. Destino de todas as Notificações daquele Bolão. Um Bolão tem no máximo um Grupo Vinculado. O vínculo é configurado pelo Administrador via painel de administração. Notificações de pré-jogo só são enviadas para grupos de Bolões que têm Palpites registrados naquele Jogo — nunca para todos os grupos indiscriminadamente. O envio é feito por um serviço standalone (whatsmeow) chamado pelo backend a cada ciclo de sincronização (5 minutos); as janelas de disparo são [7min, 12min) antes do início para "faltam dez minutos" e [-2min, +2min) para "partida iniciando".
