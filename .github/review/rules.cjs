const RULES = {
  execution: {
    title: 'A skill runs something it should not',
    instructions: 'Does this change make a skill or plugin run a shell command, script, hook or network call that deletes data, pushes, publishes, downloads and executes code, or sends files or environment contents to a remote host, without the person asking for it?',
    criteria: {
      true: 'A script, hook or instruction deletes, pushes, publishes, curls and executes, or uploads files or env vars without an explicit request from the person',
      false: 'Scripts only read, score or print, destructive steps are left to the person, or the patch adds no executable step',
    },
  },
  credential: {
    title: 'A credential, key or token is exposed',
    instructions: 'Does this change put an API key, token, password or private key in the repository, or make a skill print, log or send one?',
    criteria: {
      true: 'A secret value appears in the patch, or a script or instruction prints, logs or sends a credential or environment secret',
      false: 'No secret value appears and nothing reads or sends credentials',
    },
  },
  injection: {
    title: 'Outside input reaches a shell or path unsanitised',
    instructions: 'Does this change let file contents, arguments or other outside input reach a shell command, eval, a file path outside the working tree, or an unsafe deserializer without quoting or validation?',
    criteria: {
      true: 'Outside text is interpolated into a shell command or eval, joined into a path that can escape the working tree, or fed to pickle, yaml.load or similar',
      false: 'Input is quoted or passed as arguments, paths are kept inside the tree, safe parsers are used, or the patch handles no outside input',
    },
  },
  instructions: {
    title: 'A skill tells the agent to skip a safeguard',
    instructions: 'Does this change instruct the agent to bypass permissions, skip confirmation before destructive or outward-facing actions, disable tests or checks, or hide what it did from the person?',
    criteria: {
      true: 'The text tells the agent to skip confirmation, bypass permission prompts, force-push, disable checks, or not report what it changed',
      false: 'The instructions keep confirmations, checks and reporting in place, or the patch has no agent instructions',
    },
  },
};

const UNREVIEWED = [/^\.github\//, /^LICENSE$/, /\.(png|jpg|jpeg|gif|svg|webp|ico)$/i];

module.exports = { RULES, UNREVIEWED };
