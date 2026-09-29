import type { Lang } from "@/lib/i18n"

// Inline `code` is written with backticks and rendered by the Manual component.
export type Block =
  | { p: string }
  | { ul: string[] }
  | { ol: string[] }
  | { code: string }
  | { note: string }
  | { table: { head: string[]; rows: string[][] } }

export interface Section {
  id: string
  title: string
  blocks: Block[]
}

export interface Manual {
  title: string
  subtitle: string
  toc: string
  sections: Section[]
}

const pt: Manual = {
  title: "Manual do Safora",
  subtitle: "Como configurar, executar e acompanhar seus backups, do primeiro acesso à rotina diária.",
  toc: "Conteúdo",
  sections: [
    {
      id: "concepts",
      title: "Conceitos",
      blocks: [
        { p: "O Safora copia pastas de um lugar para outro de forma automática. Toda a configuração fica em tarefas. Estes são os termos usados no painel:" },
        {
          table: {
            head: ["Termo", "O que é"],
            rows: [
              ["Tarefa", "Uma rotina de backup: o que copiar, para onde, quando e por quanto tempo guardar."],
              ["Origem", "A pasta que será copiada. Uma tarefa pode ter várias origens."],
              ["Destino", "A pasta onde a cópia será guardada: outro disco, um disco externo ou uma pasta de rede."],
              ["Execução", "Cada vez que uma tarefa roda. Fica registrada no Histórico, com resultado e detalhes."],
              ["Agendamento", "O horário em que a tarefa roda sozinha."],
              ["Retenção", "A regra que apaga cópias antigas para o destino não encher."],
            ],
          },
        },
      ],
    },
    {
      id: "running",
      title: "Como iniciar o Safora",
      blocks: [
        { p: "O Safora precisa estar em execução para funcionar. Ele pode rodar de dois modos:" },
        {
          table: {
            head: ["Modo", "Comando", "O que permite"],
            rows: [
              ["Somente painel", "`safora serve`", "Criar tarefas e executá-las com o botão Executar agora. Não roda agendamentos."],
              ["Completo", "`safora service run`", "Tudo do modo anterior, mais agendamentos e execução automática ao conectar um disco externo."],
            ],
          },
        },
        { p: "No Windows, o instalador (`Safora-Setup`) já cadastra o Safora como serviço do computador e cria um atalho para o painel no menu Iniciar: basta executá-lo como administrador. No Linux, extraia o pacote `.tar.gz` e execute `sudo ./install.sh`. Nos dois casos o Safora passa a iniciar sozinho com o computador, no modo completo." },
        { p: "Sem instalador, também é possível cadastrar o serviço manualmente com `safora service install` e `safora service start`. Esses comandos exigem permissão de administrador; peça ajuda a quem administra o computador, se necessário." },
        { p: "Depois de iniciar, abra `http://127.0.0.1:3434` no navegador. O painel só abre no próprio computador onde o Safora está instalado." },
        { p: "As tarefas e o histórico ficam guardados em uma pasta de dados do sistema: `C:\\ProgramData\\Safora` no Windows e `/var/lib/safora` no Linux. Ao desinstalar, essa pasta é mantida, e reinstalar recupera as tarefas. Para guardar os dados ao lado do programa, por exemplo em um pendrive, inicie o Safora com a opção `--portable`." },
        { note: "O modo portátil e o modo de sistema usam bancos separados. Use sempre o mesmo modo, senão as tarefas parecem ter sumido." },
      ],
    },
    {
      id: "first-job",
      title: "Sua primeira tarefa",
      blocks: [
        { p: "Exemplo: copiar a pasta `C:\\Dados\\Projetos` para um disco externo `E:\\`, todos os dias às 02:00, guardando as últimas 7 cópias." },
        {
          ol: [
            "Abra a aba Tarefas e clique em Novo backup.",
            "Na aba Geral, digite um nome que explique a tarefa, por exemplo Projetos - diário. Em Agendamento, escolha Todos os dias e defina 02:00.",
            "Na aba Origem e destino, em Copiar de, digite `C:\\Dados\\Projetos`. Você também pode clicar no botão de pasta para escolher navegando.",
            "Em Salvar em, digite `E:\\Backups\\Projetos\\{today}`. A variável `{today}` cria uma pasta nova por dia, com a data (por exemplo `2026-09-29`).",
            "Na aba Opções, escolha Manter as últimas 7 cópias. Se os dados forem muito importantes, ative Verificar integridade.",
            "Clique em Criar tarefa. O card mostra a origem, o destino, o agendamento e a retenção.",
            "Antes de confiar no agendamento, clique em Executar agora e acompanhe a Telemetria ao vivo no Painel. Depois veja o resultado na aba Histórico.",
            "Abra a pasta de destino no computador e confirme que os arquivos estão lá.",
          ],
        },
        { p: "Só considere a tarefa pronta depois de uma execução manual bem-sucedida e de conferir os arquivos copiados." },
      ],
    },
    {
      id: "paths",
      title: "Origem, destino e variáveis de data",
      blocks: [
        { p: "Origem e destino são pastas do computador onde o Safora está instalado. Use o botão de pasta ao lado do campo para navegar até elas. Discos externos e pastas de rede precisam estar conectados ao computador; o Safora não conecta por você." },
        { p: "Nos caminhos você pode usar variáveis que o Safora troca pelo valor do dia em que a tarefa roda:" },
        {
          table: {
            head: ["Variável", "Exemplo de resultado", "Para que serve"],
            rows: [
              ["`{today}`", "`2026-09-29`", "Data de hoje."],
              ["`{yesterday}`", "`2026-09-28`", "Data de ontem. Útil para copiar o arquivo que outro sistema gerou no dia anterior."],
              ["`{today:DD-MM-YYYY}`", "`29-09-2026`", "Data de hoje em outro formato. Use `DD` para dia, `MM` para mês, `YYYY` para ano com 4 dígitos e `YY` com 2."],
              ["`{offset:-7}`", "`2026-09-22`", "Data com deslocamento em dias, aqui 7 dias atrás."],
              ["`{hostname}`", "nome do computador", "Separa os backups de vários computadores na mesma pasta."],
            ],
          },
        },
        { p: "Se digitar uma variável que não existe, a tarefa acusa erro: na origem, termina com alerta; no destino, falha." },
        { p: "O que muda ao usar a data no destino:" },
        {
          ul: [
            "Com `{today}` no destino, cada dia gera uma pasta nova. Toda execução copia tudo, e a retenção consegue apagar as pastas antigas.",
            "Sem variável de data, toda execução usa a mesma pasta e só copia o que mudou, o que é mais rápido. Nesse caso não há pastas antigas para a retenção apagar.",
          ],
        },
        { p: "Várias origens e destinos:" },
        {
          ul: [
            "Cada origem é copiada para cada destino.",
            "Com uma origem, os arquivos ficam direto na pasta de destino.",
            "Com duas ou mais origens, cada uma vai para uma subpasta com o nome da pasta de origem (por exemplo `Projetos` e `Documentos`). Nomes repetidos recebem um número (`Projetos-2`).",
          ],
        },
        { note: "Nunca escolha um destino que esteja dentro da origem. A cada execução o Safora copiaria também as cópias anteriores, e o espaço ocupado cresceria sem parar." },
      ],
    },
    {
      id: "exclusions",
      title: "Ignorar pastas e arquivos",
      blocks: [
        { p: "Em cada origem é possível informar o que não deve ser copiado. Separe os itens por vírgula:" },
        {
          table: {
            head: ["Campo", "Como funciona", "Exemplo"],
            rows: [
              ["Pastas a ignorar", "Ignora toda pasta com esse nome, em qualquer lugar dentro da origem.", "`node_modules, temp, cache`"],
              ["Arquivos a ignorar", "Ignora arquivos pelo nome. O asterisco (`*`) vale por qualquer trecho do nome.", "`*.tmp, *.log, Thumbs.db`"],
            ],
          },
        },
        {
          ul: [
            "A regra olha só o nome, não o caminho. `temp` ignora qualquer pasta chamada temp, em qualquer nível.",
            "O que é ignorado não é copiado e também não é apagado do destino pela opção Espelhar exclusões.",
            "Para ignorar coisas diferentes em origens diferentes, defina as regras em cada origem.",
          ],
        },
      ],
    },
    {
      id: "schedule",
      title: "Agendamento",
      blocks: [
        {
          table: {
            head: ["Opção", "Como usar"],
            rows: [
              ["Somente manual", "A tarefa só roda quando você clica em Executar agora ou quando conecta o disco de destino."],
              ["A cada alguns minutos", "Repete a cada 5, 10, 15, 20 ou 30 minutos."],
              ["A cada algumas horas", "Repete a cada 1, 2, 3, 4, 6, 8 ou 12 horas."],
              ["Todos os dias", "Roda uma vez por dia, no horário escolhido."],
              ["Dias da semana específicos", "Escolha os dias (por exemplo segunda, quarta e sexta) e o horário."],
              ["Uma vez por mês", "Escolha o dia do mês (1 a 28) e o horário."],
              ["Avançado", "Para quem já conhece o formato cron. Escreva a expressão diretamente."],
            ],
          },
        },
        {
          ul: [
            "Os horários seguem o relógio do computador onde o Safora está instalado.",
            "Alterar ou remover o agendamento vale na hora; não é preciso reiniciar nada.",
            "Se o computador estiver desligado no horário marcado, aquela execução não acontece e não é recuperada depois.",
            "Uma tarefa nunca roda duas vezes ao mesmo tempo. Se um novo início ocorrer durante uma execução em andamento, ele é ignorado e aparece um aviso na telemetria.",
            "Agendamentos só funcionam com o Safora rodando no modo completo ou como serviço.",
          ],
        },
        { p: "Evite marcar várias tarefas pesadas para o mesmo horário no mesmo disco. Espalhe os horários." },
      ],
    },
    {
      id: "drives",
      title: "Executar ao conectar um disco externo",
      blocks: [
        { p: "Se o destino da tarefa está em um disco externo, o Safora pode executá-la sozinho quando você conectar o disco. Ele verifica a cada 10 segundos se o disco apareceu." },
        {
          ul: [
            "No Windows, vale para destinos que começam com a letra da unidade, como `E:\\Backups`.",
            "No Linux, vale para destinos dentro de `/media/seu-usuario/nome-do-disco` ou `/mnt/nome-do-disco`.",
            "Se o disco já estava conectado quando o Safora iniciou, a tarefa não roda por isso. Só uma conexão feita depois dispara a execução.",
            "Desconectar e conectar de novo dispara a tarefa outra vez.",
            "Funciona no modo completo ou como serviço.",
            "Você pode combinar com um agendamento na mesma tarefa: os dois funcionam de forma independente.",
          ],
        },
      ],
    },
    {
      id: "retention",
      title: "Retenção: apagar cópias antigas",
      blocks: [
        { p: "Sem retenção, o destino enche. Depois de cada execução, o Safora pode apagar as cópias antigas. Você escolhe:" },
        {
          ul: [
            "Nunca apagar (padrão em tarefas novas): guarda todas as cópias. É a escolha segura; você apaga manualmente quando precisar de espaço.",
            "Manter as últimas N cópias (3, 5, 10 ou 30): guarda as N pastas mais recentes.",
            "Manter os últimos N dias (7, 30 ou 90): apaga as pastas mais antigas que N dias.",
          ],
        },
        { p: "Para funcionar, o destino precisa ter uma variável de data. Por exemplo, em `E:\\Backups\\Projetos\\{today}` o Safora entende que cada pasta dentro de `E:\\Backups\\Projetos\\` é uma cópia e apaga as mais antigas." },
        { p: "Medidas de segurança:" },
        {
          ul: [
            "Se a execução atual falhou, nada é apagado.",
            "A cópia mais recente nunca é apagada.",
            "Se só existe uma pasta, nada é apagado.",
          ],
        },
        { note: "A retenção considera todas as pastas que estiverem dentro de `E:\\Backups\\Projetos\\`, mesmo as que você criou por conta própria. Use uma pasta exclusiva para cada tarefa e não guarde outros arquivos nela." },
        { p: "Sem variável de data no destino, a retenção não age. O formulário mostra um aviso quando isso acontece." },
      ],
    },
    {
      id: "options",
      title: "Opções de cópia",
      blocks: [
        { p: "Como cada execução se comporta:" },
        {
          ul: [
            "Arquivos que não mudaram desde a última cópia são pulados. Só o que é novo ou foi alterado é copiado, e os arquivos alterados substituem a versão antiga.",
            "Arquivos apagados na origem continuam no destino, a menos que você ligue Espelhar exclusões.",
            "Se um arquivo não puder ser copiado (por exemplo, está aberto em outro programa), o Safora tenta de novo, esperando entre as tentativas. Por padrão são 3 tentativas com 30 segundos de espera; ambos os valores podem ser mudados na aba Opções. Se ainda falhar, copia o resto e termina a execução com alerta.",
          ],
        },
        {
          table: {
            head: ["Opção", "O que faz", "Quando usar"],
            rows: [
              ["Verificar integridade", "Depois de copiar cada arquivo, compara a cópia com o original. Se forem diferentes, registra o erro e marca a execução com alerta.", "Dados importantes. Deixa a execução mais lenta, porque cada arquivo copiado é lido duas vezes."],
              ["Espelhar exclusões", "No final, apaga do destino o que já não existe na origem, para o destino ficar igual à origem.", "Quando o destino deve ser um reflexo da origem. Vem desligada porque apaga arquivos do backup."],
              ["Ativar VSS", "Serviria para copiar arquivos em uso no Windows. Ainda não está disponível; o controle fica desabilitado.", "Não utilizável nesta versão."],
            ],
          },
        },
        { p: "Espelhar exclusões só apaga com segurança: se algum arquivo falhou, se a origem está vazia ou se não foi possível ler a origem, nada é apagado e um aviso é registrado. Cada item apagado aparece no log da execução. Não faz sentido usar essa opção junto com uma variável de data no destino, pois a pasta do dia começa vazia." },
      ],
    },
    {
      id: "runs",
      title: "Acompanhando as execuções",
      blocks: [
        {
          table: {
            head: ["Resultado", "Significado"],
            rows: [
              ["Sucesso", "Tudo o que precisava ser copiado foi copiado, ou já estava atualizado."],
              ["Alerta", "A execução terminou, mas com um problema parcial: uma variável inválida na origem, um arquivo que não pôde ser copiado ou uma diferença encontrada na verificação. Abra o log para ver."],
              ["Falhou", "Algo impediu a cópia, como uma origem que não existe, sem permissão de leitura, ou um destino que não pôde ser criado. Uma execução interrompida porque o Safora foi reiniciado também aparece como falha."],
              ["Cancelada", "Você interrompeu a execução pelo botão Cancelar execução. O que já foi copiado permanece, e nenhuma cópia antiga é apagada pela retenção."],
              ["Em execução", "Ainda em andamento."],
            ],
          },
        },
        { p: "Telemetria ao vivo (aba Painel): mostra em tempo real o que está sendo feito, com horário, nível (INFO, WARNING, ERROR) e mensagem. Ela é limpa sozinha quando uma nova execução começa, e o botão Limpar esvazia manualmente. O indicador Ao vivo mostra se o painel está conectado ao Safora." },
        { p: "Cancelar: enquanto uma tarefa está em execução, o botão Executar agora do card vira Cancelar execução. A execução para em instantes, inclusive no meio de um arquivo grande ou de uma espera por arquivo travado. Sem esse botão, uma tarefa presa em arquivos bloqueados só terminaria ao reiniciar o Safora." },
        { p: "Histórico: lista as 100 execuções mais recentes. Filtre por resultado e por tarefa, e clique em uma linha para ver o log completo daquela execução. Ao terminar, o Safora também pode mostrar uma notificação na área de trabalho." },
      ],
    },
    {
      id: "organize",
      title: "Organizando suas tarefas",
      blocks: [
        { p: "Boas práticas para manter tudo previsível:" },
        {
          ul: [
            "Crie uma tarefa para cada grupo de dados com a mesma importância e a mesma frequência. Não misture arquivos essenciais com descartáveis: verificação e retenção valem para a tarefa inteira.",
            "Dê nomes que digam o quê, quando e para onde, por exemplo Financeiro - diário - NAS.",
            "Guarde o backup em outro disco. Uma cópia no mesmo disco não protege se o disco quebrar.",
            "Para dados importantes, mantenha pelo menos duas cópias em lugares diferentes: dois destinos na mesma tarefa ou duas tarefas.",
            "Use uma pasta exclusiva no destino para cada tarefa, principalmente com retenção.",
            "Espalhe os horários das tarefas que gravam no mesmo disco.",
            "Sempre que mudar origem, destino ou regras de exclusão, execute manualmente uma vez e confira o resultado.",
            "Olhe o Histórico com frequência. Uma tarefa que termina sempre em alerta precisa de atenção.",
          ],
        },
        { p: "Exemplos de configuração:" },
        {
          table: {
            head: ["Situação", "Configuração sugerida"],
            rows: [
              ["Documentos pessoais, cópia diária", "Origem `C:\\Users\\Nome\\Documentos`; destino `E:\\Backups\\Documentos\\{today}`; todos os dias às 02:00; manter as últimas 7 cópias; verificar integridade."],
              ["Arquivo gerado todo dia por outro sistema", "Origem `D:\\Exports\\{yesterday:DD-MM-YYYY}`; destino `Z:\\backup\\exports\\{today}` (pasta de rede conectada como Z:); todos os dias às 06:00; manter 30 dias."],
              ["Disco externo que você conecta de tempos em tempos", "Destino `E:\\Backups\\Projetos` sem variável de data e sem agendamento. A tarefa roda ao conectar o disco e copia só o que mudou."],
              ["Servidor Linux com várias pastas", "Duas origens (`/srv/app/dados` e `/etc/app`); destino `/mnt/backup/app/{today}`; todos os dias às 03:30. Cada origem vai para a sua subpasta."],
            ],
          },
        },
      ],
    },
    {
      id: "import",
      title: "Importando scripts antigos",
      blocks: [
        { p: "Se você já faz backup com um script do Windows (`.bat` com `robocopy`), não precisa recriar tudo. Em Tarefas, clique em Importar script e cole o conteúdo. O Safora identifica origem, destino, pastas e arquivos ignorados, número de tentativas e variáveis de data, e cria uma tarefa chamada Imported Job, sem apagar cópias antigas. Scripts que montam a data em partes (DD, MM e YY em variáveis separadas) e comandos com `/MIR` também são reconhecidos: `/MIR` liga a opção Espelhar exclusões." },
        {
          ul: [
            "Abra a tarefa criada, dê um nome melhor e defina o agendamento.",
            "Revise origem, destino e exclusões. Se o script usava uma data no caminho, a variável equivalente aparece no campo.",
            "Execute manualmente e confira o resultado antes de desligar o script antigo.",
          ],
        },
      ],
    },
    {
      id: "troubleshooting",
      title: "Solução de problemas",
      blocks: [
        {
          table: {
            head: ["Problema", "Causa provável", "O que fazer"],
            rows: [
              ["A tarefa não rodou no horário", "O Safora estava no modo somente painel, ou o computador estava desligado.", "Use o modo completo ou instale o serviço. Confira o agendamento no card da tarefa."],
              ["Aparece Desconectado", "O Safora parou ou a página perdeu a conexão.", "Inicie o Safora de novo e recarregue a página."],
              ["Ao clicar em Executar agora, a tarefa já está em execução", "Há uma execução dessa tarefa em andamento.", "Aguarde terminar. Acompanhe na telemetria."],
              ["Resultado Alerta", "Variável inválida, arquivo que não copiou ou diferença na verificação.", "Abra a execução no Histórico e leia as linhas WARNING e ERROR."],
              ["Resultado Falhou", "A origem não existe, não há permissão de leitura ou o disco não está conectado.", "Confira se a pasta existe e se o disco está conectado. Verifique as permissões do usuário que roda o Safora."],
              ["A retenção não apaga nada", "O destino não tem variável de data, a execução falhou ou só existe uma pasta.", "Inclua `{today}` no destino e leia a seção de retenção."],
              ["A retenção apagou pastas que eu usava para outra coisa", "A pasta principal do destino tinha outros dados.", "Use uma pasta exclusiva por tarefa."],
              ["Conectei o disco e nada rodou", "O disco já estava conectado quando o Safora iniciou, ou o destino não está em um caminho de disco externo.", "Desconecte e conecte de novo. Confira o formato do destino na seção sobre discos externos."],
              ["Permissão negada", "O usuário que roda o Safora não consegue ler a origem ou gravar no destino.", "Ajuste as permissões das pastas."],
              ["Arquivos indesejados no backup", "Falta uma regra de exclusão, ou o nome não bate.", "Revise as regras. Elas comparam só o nome, não o caminho."],
              ["Minhas tarefas sumiram", "O Safora foi iniciado a partir de outra pasta.", "Inicie a partir da pasta que contém o safora.db original."],
            ],
          },
        },
      ],
    },
    {
      id: "limits",
      title: "Limitações desta versão",
      blocks: [
        {
          ul: [
            "Os destinos são pastas do computador. Não há envio direto para serviços de nuvem; use uma pasta sincronizada ou uma unidade de rede conectada.",
            "Não existe função de restauração no painel. As cópias são arquivos comuns: para recuperar, copie os arquivos de volta pelo gerenciador de arquivos.",
            "O VSS (cópia de arquivos em uso no Windows) ainda não está disponível.",
            "O Histórico mostra as 100 execuções mais recentes.",
            "O painel só abre no computador onde o Safora está instalado.",
          ],
        },
      ],
    },
  ],
}

const en: Manual = {
  title: "Safora Manual",
  subtitle: "How to set up, run and follow your backups, from the first access to daily routine.",
  toc: "Contents",
  sections: [
    {
      id: "concepts",
      title: "Concepts",
      blocks: [
        { p: "Safora copies folders from one place to another automatically. All configuration lives in jobs. These are the terms used in the dashboard:" },
        {
          table: {
            head: ["Term", "What it is"],
            rows: [
              ["Job", "A backup routine: what to copy, where to, when, and how long to keep it."],
              ["Source", "The folder to be copied. A job can have several sources."],
              ["Destination", "The folder where the copy is stored: another disk, an external drive or a network folder."],
              ["Run", "Each time a job executes. It is recorded in History, with its result and details."],
              ["Schedule", "The time at which the job runs by itself."],
              ["Retention", "The rule that deletes old copies so the destination does not fill up."],
            ],
          },
        },
      ],
    },
    {
      id: "running",
      title: "Starting Safora",
      blocks: [
        { p: "Safora must be running to work. It can run in two modes:" },
        {
          table: {
            head: ["Mode", "Command", "What it allows"],
            rows: [
              ["Dashboard only", "`safora serve`", "Create jobs and run them with the Run Now button. Does not run schedules."],
              ["Full", "`safora service run`", "Everything in the previous mode, plus schedules and automatic runs when an external drive is connected."],
            ],
          },
        },
        { p: "On Windows, the installer (`Safora-Setup`) already registers Safora as a computer service and creates a Start menu shortcut to the dashboard: just run it as administrator. On Linux, extract the `.tar.gz` package and run `sudo ./install.sh`. In both cases Safora starts by itself with the computer, in full mode." },
        { p: "Without the installer, you can also register the service manually with `safora service install` and `safora service start`. These commands require administrator permission; ask whoever manages the computer if needed." },
        { p: "After starting, open `http://127.0.0.1:3434` in your browser. The dashboard only opens on the computer where Safora is installed." },
        { p: "Jobs and history are stored in a system data folder: `C:\\ProgramData\\Safora` on Windows and `/var/lib/safora` on Linux. Uninstalling keeps that folder, and reinstalling recovers your jobs. To keep the data next to the program, for example on a USB stick, start Safora with the `--portable` option." },
        { note: "Portable mode and system mode use separate databases. Always use the same mode, otherwise your jobs seem to have disappeared." },
      ],
    },
    {
      id: "first-job",
      title: "Your first job",
      blocks: [
        { p: "Example: copy the folder `C:\\Data\\Projects` to an external drive `E:\\`, every day at 02:00, keeping the last 7 copies." },
        {
          ol: [
            "Open the Jobs tab and click New Backup.",
            "On the General tab, type a name that explains the job, for example Projects - daily. Under Schedule, choose Every day and set 02:00.",
            "On the Source & destination tab, under Back up from, type `C:\\Data\\Projects`. You can also click the folder button to browse.",
            "Under Save to, type `E:\\Backups\\Projects\\{today}`. The `{today}` variable creates a new folder per day, named with the date (for example `2026-09-29`).",
            "On the Options tab, choose Keep last 7 copies. If the data is very important, turn on Verify integrity.",
            "Click Create Job. The card shows the source, destination, schedule and retention.",
            "Before trusting the schedule, click Run Now and follow the Live Telemetry on the Dashboard. Then check the result in the History tab.",
            "Open the destination folder on the computer and confirm the files are there.",
          ],
        },
        { p: "Only consider the job ready after a successful manual run and after checking the copied files." },
      ],
    },
    {
      id: "paths",
      title: "Source, destination and date variables",
      blocks: [
        { p: "Source and destination are folders on the computer where Safora is installed. Use the folder button next to the field to browse to them. External drives and network folders must be connected to the computer; Safora does not connect them for you." },
        { p: "In paths you can use variables that Safora replaces with the value of the day the job runs:" },
        {
          table: {
            head: ["Variable", "Example result", "What it is for"],
            rows: [
              ["`{today}`", "`2026-09-29`", "Today's date."],
              ["`{yesterday}`", "`2026-09-28`", "Yesterday's date. Useful to copy the file another system generated the day before."],
              ["`{today:DD-MM-YYYY}`", "`29-09-2026`", "Today's date in another format. Use `DD` for day, `MM` for month, `YYYY` for a 4-digit year and `YY` for 2 digits."],
              ["`{offset:-7}`", "`2026-09-22`", "Date shifted by days, here 7 days ago."],
              ["`{hostname}`", "computer name", "Separates the backups of several computers in the same folder."],
            ],
          },
        },
        { p: "If you type a variable that does not exist, the job reports an error: in the source, it finishes with a warning; in the destination, it fails." },
        { p: "What changes when you use the date in the destination:" },
        {
          ul: [
            "With `{today}` in the destination, each day creates a new folder. Every run copies everything, and retention can delete the old folders.",
            "Without a date variable, every run uses the same folder and only copies what changed, which is faster. In that case there are no old folders for retention to delete.",
          ],
        },
        { p: "Several sources and destinations:" },
        {
          ul: [
            "Each source is copied to each destination.",
            "With one source, the files go straight into the destination folder.",
            "With two or more sources, each goes into a subfolder named after the source folder (for example `Projects` and `Documents`). Repeated names get a number (`Projects-2`).",
          ],
        },
        { note: "Never choose a destination that is inside the source. On each run Safora would also copy the previous copies, and the space used would grow without limit." },
      ],
    },
    {
      id: "exclusions",
      title: "Skipping folders and files",
      blocks: [
        { p: "On each source you can list what must not be copied. Separate items with commas:" },
        {
          table: {
            head: ["Field", "How it works", "Example"],
            rows: [
              ["Folders to skip", "Skips every folder with that name, anywhere inside the source.", "`node_modules, temp, cache`"],
              ["Files to skip", "Skips files by name. The asterisk (`*`) stands for any part of the name.", "`*.tmp, *.log, Thumbs.db`"],
            ],
          },
        },
        {
          ul: [
            "The rule looks only at the name, not the path. `temp` skips any folder called temp, at any level.",
            "Skipped items are not copied and are also not deleted from the destination by Mirror deletions.",
            "To skip different things in different sources, set the rules on each source.",
          ],
        },
      ],
    },
    {
      id: "schedule",
      title: "Scheduling",
      blocks: [
        {
          table: {
            head: ["Option", "How to use it"],
            rows: [
              ["Manual only", "The job only runs when you click Run Now or when you connect the destination drive."],
              ["Every few minutes", "Repeats every 5, 10, 15, 20 or 30 minutes."],
              ["Every few hours", "Repeats every 1, 2, 3, 4, 6, 8 or 12 hours."],
              ["Every day", "Runs once a day, at the chosen time."],
              ["Certain days of the week", "Choose the days (for example Monday, Wednesday and Friday) and the time."],
              ["Once a month", "Choose the day of the month (1 to 28) and the time."],
              ["Advanced", "For those who already know the cron format. Type the expression directly."],
            ],
          },
        },
        {
          ul: [
            "Times follow the clock of the computer where Safora is installed.",
            "Changing or removing the schedule applies immediately; nothing needs to be restarted.",
            "If the computer is off at the scheduled time, that run does not happen and is not made up later.",
            "A job never runs twice at the same time. If a new start happens during a run in progress, it is ignored and a notice appears in the telemetry.",
            "Schedules only work with Safora running in full mode or as a service.",
          ],
        },
        { p: "Avoid scheduling several heavy jobs for the same time on the same disk. Spread the times out." },
      ],
    },
    {
      id: "drives",
      title: "Running when an external drive is connected",
      blocks: [
        { p: "If the job's destination is on an external drive, Safora can run it by itself when you connect the drive. It checks every 10 seconds whether the drive has appeared." },
        {
          ul: [
            "On Windows, it applies to destinations that start with the drive letter, such as `E:\\Backups`.",
            "On Linux, it applies to destinations inside `/media/your-user/disk-name` or `/mnt/disk-name`.",
            "If the drive was already connected when Safora started, the job does not run because of that. Only a connection made afterwards triggers the run.",
            "Disconnecting and connecting again triggers the job once more.",
            "Works in full mode or as a service.",
            "You can combine it with a schedule on the same job: the two work independently.",
          ],
        },
      ],
    },
    {
      id: "retention",
      title: "Retention: deleting old copies",
      blocks: [
        { p: "Without retention, the destination fills up. After each run, Safora can delete old copies. You choose:" },
        {
          ul: [
            "Never delete (default for new jobs): keeps every copy. This is the safe choice; you delete manually when you need space.",
            "Keep last N copies (3, 5, 10 or 30): keeps the N most recent folders.",
            "Keep last N days (7, 30 or 90): deletes folders older than N days.",
          ],
        },
        { p: "To work, the destination must have a date variable. For example, in `E:\\Backups\\Projects\\{today}` Safora understands that each folder inside `E:\\Backups\\Projects\\` is a copy and deletes the oldest ones." },
        { p: "Safety measures:" },
        {
          ul: [
            "If the current run failed, nothing is deleted.",
            "The most recent copy is never deleted.",
            "If there is only one folder, nothing is deleted.",
          ],
        },
        { note: "Retention considers every folder inside `E:\\Backups\\Projects\\`, even ones you created yourself. Use an exclusive folder for each job and do not keep other files in it." },
        { p: "Without a date variable in the destination, retention does nothing. The form shows a warning when that happens." },
      ],
    },
    {
      id: "options",
      title: "Copy options",
      blocks: [
        { p: "How each run behaves:" },
        {
          ul: [
            "Files that have not changed since the last copy are skipped. Only what is new or changed is copied, and changed files replace the old version.",
            "Files deleted from the source stay in the destination unless you turn on Mirror deletions.",
            "If a file cannot be copied (for example, it is open in another program), Safora retries, waiting between attempts. By default that is 3 retries with a 30 second wait; both values can be changed on the Options tab. If it still fails, it copies the rest and ends the run with a warning.",
          ],
        },
        {
          table: {
            head: ["Option", "What it does", "When to use it"],
            rows: [
              ["Verify integrity", "After copying each file, compares the copy with the original. If they differ, it logs the error and marks the run with a warning.", "Important data. Makes the run slower, because every copied file is read twice."],
              ["Mirror deletions", "At the end, deletes from the destination what no longer exists in the source, so the destination matches the source.", "When the destination should mirror the source. Off by default because it deletes backup files."],
              ["Enable VSS", "Would copy files in use on Windows. Not available yet; the control is disabled.", "Not usable in this version."],
            ],
          },
        },
        { p: "Mirror deletions only deletes safely: if any file failed, if the source is empty or if the source could not be read, nothing is deleted and a warning is logged. Each deleted item appears in the run log. It makes no sense to use this option with a date variable in the destination, since the day's folder starts empty." },
      ],
    },
    {
      id: "runs",
      title: "Following your runs",
      blocks: [
        {
          table: {
            head: ["Result", "Meaning"],
            rows: [
              ["Success", "Everything that needed copying was copied, or was already up to date."],
              ["Warning", "The run finished, but with a partial problem: an invalid variable in the source, a file that could not be copied or a difference found during verification. Open the log to see."],
              ["Failed", "Something prevented the copy, such as a source that does not exist, no read permission, or a destination that could not be created. A run interrupted because Safora was restarted also shows as failed."],
              ["Cancelled", "You stopped the run with the Cancel run button. What was already copied stays, and retention deletes no old copies."],
              ["Running", "Still in progress."],
            ],
          },
        },
        { p: "Live telemetry (Dashboard tab): shows in real time what is being done, with time, level (INFO, WARNING, ERROR) and message. It clears itself when a new run starts, and the Clear button empties it manually. The Live indicator shows whether the dashboard is connected to Safora." },
        { p: "Cancelling: while a job is running, the Run Now button on its card becomes Cancel run. The run stops within moments, even in the middle of a large file or while waiting to retry a locked file. Without this button, a job stuck on locked files would only end by restarting Safora." },
        { p: "History: lists the 100 most recent runs. Filter by result and by job, and click a row to see the full log of that run. When it finishes, Safora may also show a desktop notification." },
      ],
    },
    {
      id: "organize",
      title: "Organizing your jobs",
      blocks: [
        { p: "Good practices to keep everything predictable:" },
        {
          ul: [
            "Create one job for each group of data with the same importance and frequency. Do not mix essential and disposable files: verification and retention apply to the whole job.",
            "Use names that say what, when and where to, for example Finance - daily - NAS.",
            "Store the backup on a different disk. A copy on the same disk does not help if the disk breaks.",
            "For important data, keep at least two copies in different places: two destinations in the same job or two jobs.",
            "Use an exclusive folder in the destination for each job, especially with retention.",
            "Spread out the times of jobs that write to the same disk.",
            "Whenever you change source, destination or exclusion rules, run manually once and check the result.",
            "Look at History often. A job that always ends with a warning needs attention.",
          ],
        },
        { p: "Configuration examples:" },
        {
          table: {
            head: ["Situation", "Suggested configuration"],
            rows: [
              ["Personal documents, daily copy", "Source `C:\\Users\\Name\\Documents`; destination `E:\\Backups\\Documents\\{today}`; every day at 02:00; keep last 7 copies; verify integrity."],
              ["File generated every day by another system", "Source `D:\\Exports\\{yesterday:DD-MM-YYYY}`; destination `Z:\\backup\\exports\\{today}` (network folder connected as Z:); every day at 06:00; keep 30 days."],
              ["External drive you connect from time to time", "Destination `E:\\Backups\\Projects` without a date variable and without a schedule. The job runs when the drive is connected and copies only what changed."],
              ["Linux server with several folders", "Two sources (`/srv/app/data` and `/etc/app`); destination `/mnt/backup/app/{today}`; every day at 03:30. Each source goes into its own subfolder."],
            ],
          },
        },
      ],
    },
    {
      id: "import",
      title: "Importing old scripts",
      blocks: [
        { p: "If you already back up with a Windows script (a `.bat` using `robocopy`), you do not need to rebuild everything. In Jobs, click Import Script and paste the content. Safora identifies the source, destination, skipped folders and files, number of retries and date variables, and creates a job called Imported Job that does not delete old copies. Scripts that build the date from separate parts (DD, MM and YY in separate variables) and commands with `/MIR` are also recognized: `/MIR` turns on Mirror deletions." },
        {
          ul: [
            "Open the created job, give it a better name and set the schedule.",
            "Review source, destination and exclusions. If the script used a date in the path, the equivalent variable appears in the field.",
            "Run it manually and check the result before turning off the old script.",
          ],
        },
      ],
    },
    {
      id: "troubleshooting",
      title: "Troubleshooting",
      blocks: [
        {
          table: {
            head: ["Problem", "Likely cause", "What to do"],
            rows: [
              ["The job did not run at the scheduled time", "Safora was in dashboard-only mode, or the computer was off.", "Use full mode or install the service. Check the schedule on the job card."],
              ["Disconnected shows up", "Safora stopped or the page lost its connection.", "Start Safora again and reload the page."],
              ["Run Now says the job is already running", "A run of that job is in progress.", "Wait for it to finish. Follow it in the telemetry."],
              ["Warning result", "Invalid variable, a file that did not copy or a difference in verification.", "Open the run in History and read the WARNING and ERROR lines."],
              ["Failed result", "The source does not exist, there is no read permission or the drive is not connected.", "Check that the folder exists and the drive is connected. Check the permissions of the user running Safora."],
              ["Retention deletes nothing", "The destination has no date variable, the run failed or there is only one folder.", "Add `{today}` to the destination and read the retention section."],
              ["Retention deleted folders I used for something else", "The destination's parent folder had other data.", "Use an exclusive folder per job."],
              ["I connected the drive and nothing ran", "The drive was already connected when Safora started, or the destination is not on an external drive path.", "Disconnect and reconnect. Check the destination format in the external drive section."],
              ["Permission denied", "The user running Safora cannot read the source or write to the destination.", "Adjust the folder permissions."],
              ["Unwanted files in the backup", "A skip rule is missing, or the name does not match.", "Review the rules. They compare only the name, not the path."],
              ["My jobs disappeared", "Safora was started from another folder.", "Start it from the folder that contains the original safora.db."],
            ],
          },
        },
      ],
    },
    {
      id: "limits",
      title: "Limitations of this version",
      blocks: [
        {
          ul: [
            "Destinations are folders on the computer. There is no direct upload to cloud services; use a synced folder or a connected network drive.",
            "There is no restore function in the dashboard. Copies are ordinary files: to recover, copy the files back with the file manager.",
            "VSS (copying files in use on Windows) is not available yet.",
            "History shows the 100 most recent runs.",
            "The dashboard only opens on the computer where Safora is installed.",
          ],
        },
      ],
    },
  ],
}

export const manuals: Record<Lang, Manual> = { pt, en }
