[
  {
    matcher: {
      id: 'byName',
      options: 'r%s' % l,
    },
    properties: [
      {
        id: 'color',
        value: {
          fixedColor: 'rgba(%s, 0, 0, 1)' % l,
          mode: 'fixed',
        },
      },
    ],
  }
  for l in std.range(0, 255)
] + [
  {
    matcher: {
      id: 'byName',
      options: 'g%s' % l,
    },
    properties: [
      {
        id: 'color',
        value: {
          fixedColor: 'rgba(0, %s, 0, 1)' % l,
          mode: 'fixed',
        },
      },
    ],
  }
  for l in std.range(0, 255)
] + [
  {
    matcher: {
      id: 'byName',
      options: 'b%s' % l,
    },
    properties: [
      {
        id: 'color',
        value: {
          fixedColor: 'rgba(0, 0, %s, 1)' % (l * 2),
          mode: 'fixed',
        },
      },
    ],
  }
  for l in std.range(0, 127)
]
