Este Readme é um modelo que pode ser usado para os repositórios no GitHub de projetos feitos para a atividade de Extensão Hackerspace, da Universidade Federal de São Carlos - Campus Sorocaba, organizada pelo hackerspace [HackoonSpace](https://hackoonspace.com/). É importante observar que alguns campos e informações são opcionais, e que não é necessário seguir rigorosametne o modelo proposto. O objetivo é que, obrigatóriamente, os grupos coloquem em seus repositórios as informações requisitadas aqui, porém o formato em si pode ou não seguir o modelo.

Além disso, para cada tópico, existem alguns possíveis exemplos de como o grupo pode apresentar as informações sobre o projeto (ex: lista enumerada, texto descritivo, etc). Alguns tópicos possuem comandos que só podem ser melhor visualizados na versão Raw ou txt do Readme.

É preferível que a linguagem da documentação esteja em português, mas também é aceito que esteja em inglês.

Para responder quaisquer dúvidas, entrar em contato com a equipe do HackoonSpace.

# Titulo do projeto

## Conceito do projeto
Aqui vocês definem o conceito do projeto, explicando sucintamente sobre o que ele se trata, qual a finalidade/utilidade, etc. 1-2 parágrafos já são suficientes, desde que bem explicados, mas sintam-se a vontade para escrever mais. Também sintam-se a vontade para usar a criatividade.

### Exemplo:

Este projeto foi desenvolvido com o intuíto de facilitar a manuntenção de aplicações Web contra invasões indesejadas do tipo XYZ. Para isto, foi implementado um programa que realizasse a checagem de pacotes recebidos por um Website com host local em uma máquina com Windows, buscando traços de possíveis códigos maliciosos que possam interferir na execução habitual do servidor da aplicação.
  
## Pré-requisitos e recursos utilizados
Citação das linguagens, bibliotecas, peças de hardware, e outras coisas que o grupo utilizou para realizar o projeto. Não é necessário explicar qual foi o uso exato de cada coisa no projeto. Bibliotecas e recursos padrões das tecnologias utilizadas não precisam ser citados (ex: stdio.h, iostream.h, etc.).

Se alguma biblioteca externa ou código de outra pessoa foi utilizado como recurso, é importante citar a fonte de onde vocês retiraram (pode ser o link no Github, ou tutorial usado como referência).

### Exemplo:

O grupo utilizou a linguagem C para desenvolver a implementação geral do projeto, além de importar as seguintes bibliotecas:
1. abcdzd.h
2. exemplo.h, disponível em [IstoEhApenasUmExemplo](https://github.com/istoehapenasumexemplo/minhabiblioteca)

Também foi utilizado o tutorial disponível em [IstoEhOutroExemplo](https://github.com/istoehoutroexemplo/oi) como base para o grupo compreender a implementação da função X dentro da linguagem em questão.
  
## Passo a passo
Passos que o grupo realizou para criar, implementar ou projetar o projeto. É importante descrever pelo menos o mais importante para que outras pessoas compreendam como o grupo conseguiu realizar o projeto, quais as atividades feitas, etc, e possam ter meios compreender como reproduzir o projeto, se assim fosse necessário.

Se possível, é legal citar o nome dos arquivos implementados, se forem poucos. Por exemplo, se o seu projeto tiver 4 arquivos, cada um com uma função, citar o nome deles na parte do passo a passo correspondente. Se forem muitos arquivos para uma mesma coisa, não tem problema, podem deixar sem ou deixar apenas o nome da pasta.

### Exemplo:

1. Baixamos o material disponível em [Material](https://materialdeexemplodohackerspace.com.br)
2. Estudamos como o código do material anterior funciona
3. Implementamos um programa que se comunicasse com o código compreendido (comunicacao.c e comunicacao.h)
4. Implementamos uma interface gráfica para utilizar o programa de comunicação de forma mais intuitiva.

## Instalação
Passos necessários para instalar ou recriar seu projeto, se assim for necessário. A descrição dos passos não precisa ser complexa. É necessário apenas o mais importante para que outras pessoas saibam como fazê-lo.

### Exemplos:
a)
  ```
  Execute o comando X Y Z, no terminal, na pasta do projeto
  ```
b)
  1. Abra a pasta 
  2. Execute o comando A B C no terminal
  3. Compile os arquivos X, Y e Z juntos
  4. Crie um arquivo W.txt de entrada

## Execução
Passos necessários para executar, rodar ou testar seu projeto. Vocês podem seguir o mesmo modelo dos exemplos de Instalação.

## Bugs/problemas conhecidos
Lista de possíveis problemas, bugs, falhas ou comportamentos esquisitos que o grupo conheça sobre o projeto. Esta seção é importante para que outras pessoas saibam quais tipos de erros elas podem encontrar. Seria legal citar motivos que o grupo acredita que sejam os causadores destas coisas, mas não é obrigatório.

### Exemplo:

O projeto possui uma falha ao abrir a aba INICIO, após realizar uma inserção com caractéres acentuados. Também foi encontrada uma falha ao definir a tela de fundo com a cor Roxa, provavelmente por conta da palheta de cores limitada da tecnologia que foi utilizada.

## Autores
Aqui, é importante referenciar o nome dos integrantes do grupo. Não precisa de RA. Outras informações, como contato ou perfil no Github, ficam a critério do grupo. Se o grupo for muito grande, é bom referenciar as funções de cada um.

### Exemplo:
* Marcus Vinícius N. Garcia ([Infinitemarcus](https://github.com/Infinitemarcus))
* Garcia Neto Junior da Silva
* João das Neves - Desenvolvedor do Back-End

## Demais anotações e referências (opcional)
Aqui, o grupo pode colocar quaisquer outras informações que ache relevante, se assim desejar. Links de referências e materiais de estudo utilizados ou recomendados são sempre bem vindos. 

## Imagens/screenshots
É necessário colocar pelo menos 3 imagens/screenshots do projeto, porém fiquem a vontade para colocar mais, a medida do que vocês acharem legal para ilustrar o projeto.

Para colocar imagens no Readme do Github, vocês podem usar o seguinte comando (abrir este Readme no modo raw ou como txt):

![Imagem](https://github.com/hackoonspace/Hackoonspace-template/blob/master/exemplo.png)

É preferível que vocês usem imagens hospedadas no próprio GitHub do projeto. É só referenciar o link delas no comando acima.
