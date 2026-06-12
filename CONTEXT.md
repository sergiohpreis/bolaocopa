# Glossário — Bolão Copa

## Bolão
Um grupo de apostas criado por um Administrador, compartilhado via link de convite. Contém um conjunto de Participantes, as Regras de Pontuação e os Palpites de cada Participante.

## Administrador
O Participante que criou o Bolão. Tem poderes exclusivos: gerar o link de convite e configurar as Regras de Pontuação. Existe exatamente um Administrador por Bolão.

## Participante
Qualquer usuário autenticado (via Google OAuth) que entrou no Bolão através do link de convite. Pode registrar e alterar Palpites até o início de cada Jogo.

## Jogo
Uma partida da Copa do Mundo, com dois times, horário de início e placar final. Os dados de Jogos são sincronizados automaticamente de uma API externa. O horário de início define o prazo para Palpites.

## Palpite
O placar exato que um Participante prevê para um Jogo. Pode ser criado ou alterado até o momento em que o Jogo começa. Após o início do Jogo, o Palpite é bloqueado.

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
