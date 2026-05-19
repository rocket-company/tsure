# tsure Mobile

Template inicial do app Android nativo do `tsure`.

Objetivo desta base:

- servir como ponto de partida para operacao em campo;
- permitir futuras telas de check-in, checklists, OS e apontamento de frota;
- manter uma estrutura Kotlin simples e facil de evoluir.

## Estrutura

- `app/` modulo Android principal
- `settings.gradle.kts` configuracao do projeto
- `build.gradle.kts` plugins comuns
- `gradle.properties` flags basicas do Gradle

## Abrir

Abra `apps/mobile` no Android Studio.

## Estado atual

- uma `MainActivity` basica;
- layout XML simples;
- dependencias AndroidX e Material;
- tema leve para servir de base.
