import * as path from 'node:path';
import Mocha from 'mocha';
import glob from 'glob';

async function main() {
  const mocha = new Mocha({
    ui: 'tdd',
    color: true,
  });

  const testsRoot = path.resolve(__dirname, './suite');

  return new Promise<void>((resolve, reject) => {
    glob(
      '**/*.test.js',
      { cwd: testsRoot },
      (err: Error | null, files: string[]) => {
        if (err) {
          return reject(err);
        }

        // Add files to the test suite
        files.forEach((f: string) =>
          mocha.addFile(path.resolve(testsRoot, f))
        );

        try {
          // Run the mocha test
          mocha.run((failures: number) => {
            if (failures > 0) {
              reject(new Error(`${failures} tests failed.`));
            } else {
              resolve();
            }
          });
        } catch (err) {
          console.error(err);
          reject(err);
        }
      }
    );
  });
}

main()
  .then(() => {
    console.log('All tests passed!');
    process.exit(0);
  })
  .catch((err) => {
    console.error('Failed to run tests:', err);
    process.exit(1);
  });
