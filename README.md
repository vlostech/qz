# qz

`qz` is a command line tool for self-education that provides a simple way of testing knowledge.

> [!IMPORTANT]
> At the moment the tool is under development and requires Go to be installed.

## Getting Started

Step 1 – Install the program using the following command (Go is required):

```sh
go install github.com/vlostech/qz/cmd/qz@latest
```

Step 2 – Prepare a file with questions and answers. An empty line should separate all answers and questions.

```
Question 1

Answer 1

Question 2

Answer 2
```

Step 3 – Run the program for the prepared file.

```sh
qz run -f ~/test.txt
```

You can use `-c` (`--count`) flag to specify the number of questions in a session.

You can also use `-r` (`--range`) flag to specify the range of questions that will be used in a session.

| Example | Description                                        |
| ------- | -------------------------------------------------- |
| `5`     | Question by index 5.                               |
| `..5`   | Questions from 0 inclusive to 5 exclusive.         |
| `5..`   | Questions from 5 inclusive to the end of the file. |
| `5..10` | Questions from 5 inclusive to 10 exclusive.        |
| `..`    | All questions in the file.                         |

You are able to specify multiple ranges that are separated by `,`.

```sh
qz run -f ~/test.txt -c 10 -r ..10,15,20,30..40,50..
```

If `test.txt` contains 100 questions, the example above takes 72 questions and runs a session with 10 random questions.

## Input features

When answering a question, you can use `\` character at the end of a line to add another line.

```
> First row\
> Second row\
> Last row
```

You can also use `\end` command at the beginning of a line to discard the current line.

```
> First row\
> Second row\
> Last row\
> \end
```

## Saving questions to a file

After your session is finished, you will be asked to save questions to a file.
You can choose questions from the session that you want to save using the same
range syntax as explained above. When all necessary questions are chosen, you
can proceed to provide a file path.

You can type a file path in different ways:

* Absolute path. File path starts with `/` (or `{volume}:\` for Windows}). For
  example, `/foo/bar.txt` (or `D:\foo\bar.txt` for Windows).

* Relative path to working directory. File path starts with a file name or a
  directory name. For example, `foo/bar.txt` means `{WORKDIR}/foo/bar.txt`.

* Relative path to home directory. File path starts with `~`. For example,
  `~/foo/bar.txt` means `{HOMEDIR}/foo/bar.txt`.

If an existing file is provided, questions will be appended to the file. All
duplicate questions will be skipped.

If the file does not exist, it will be created. All missing directories in the
file path will be created.
