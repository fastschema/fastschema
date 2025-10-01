'user strict';

const names = [
  'Alice',
  'Bob',
  'Charlie',
  'David',
  'Eve',
  'Frank',
  'Grace',
  'Hank',
  'Ivy',
  'Jack',
];

export const getRandomName = async () => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(names[Math.floor(Math.random() * names.length)]);
    }, 100);
  });
};
