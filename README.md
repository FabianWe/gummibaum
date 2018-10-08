# gummibaum
gummibaum is LaTeX template engine. Sometimes you just want a template with placeholders and fill those placeholders with actual content, for example using placeholders $name in your tex code and replace them with a constant "John". Or you have data in form of a csv and create LaTeX code based on that data, for example creating a dynamic table.
gummibaum tries to be simple to use: It is not a LaTeX engine like LuaTeX but a simple static content creator that produces LaTeX output that can be compiled with `pdflatex`.
It supports two different modes: Expansion mode and template mode.
# Expansion Mode
The easiest mode to understand is the expansion mode. It takes a LaTeX file as input, replaces certain placeholders with values and and can be used to iterate over specific parts of the file. The advantage is that you can write a *.tex* file that you can compile with `pdflatex` and test how it looks like. Then use gummibaum to replace constant fields or use an easy way to iterate over content. It is not as flexible as the template mode though.

## License
Copyright 2018 Fabian Wenzelmann

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
