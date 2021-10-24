# debugfile

This input takes one input file as parameter and sends this file as an event at startup and at regular intervals.
Good for debugging.

````json
{
  "input": [
    {
      "type": "debugfile",
      "input": "/etc/passwd",
      "codec": "default",
      "delay": 120
    }
  ]
}
````