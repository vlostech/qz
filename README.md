# qz

## Getting Started

1. Install the program using the following command (Go is required):

```sh
go install github.com/vlostech/qz/cmd/qz@latest
```

2. Prepare a file with questions and answers. All answers and questions should be separated by an empty line.

```
Question 1

Answer 1

Question 2

Answer 2
```

3. Run the program for the prepared file.

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
