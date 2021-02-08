workflow "New workflow" {
  on       = "push"
  resolves = [""]
}

action "Hello world" {
  uses = "./action-a"
  env = {
    MY_NAME = "B-Rad"
  }
  args = "\"Hello world, I'm $MY_NAME!\""
}
