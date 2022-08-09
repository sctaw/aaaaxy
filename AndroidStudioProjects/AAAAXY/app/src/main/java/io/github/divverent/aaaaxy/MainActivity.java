package io.github.divverent.aaaaxy;

import android.os.Bundle;
import androidx.appcompat.app.AppCompatActivity;
import androidx.core.view.WindowInsetsCompat;
import androidx.core.view.WindowInsetsControllerCompat;
import java.io.File;
import java.lang.System;
import java.lang.Thread;

import go.Seq;
import io.github.divverent.aaaaxy.aaaaxy.Aaaaxy;
import io.github.divverent.aaaaxy.aaaaxy.EbitenView;

public class MainActivity extends AppCompatActivity {
	private WindowInsetsControllerCompat insetsController;

	@Override
	protected void onCreate(Bundle savedInstanceState) {
		super.onCreate(savedInstanceState);
		Seq.setContext(getApplicationContext());
		File dir = getExternalFilesDir(null);
		Aaaaxy.setFilesDir(dir.getAbsolutePath());
		new Thread(() -> {
			Aaaaxy.waitQuit();
			finishAndRemoveTask();
			System.exit(0);
		}).start();
		setContentView(R.layout.activity_main);
		insetsController = new WindowInsetsControllerCompat(
			getWindow(), getWindow().getDecorView());
		insetsController.hide(WindowInsetsCompat.Type.systemBars());
		insetsController.setSystemBarsBehavior(
			WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE);
	}

	private EbitenView getEbitenView() {
		return (EbitenView) this.findViewById(R.id.view);
	}

	@Override
	protected void onPause() {
		super.onPause();
		this.getEbitenView().suspendGame();
	}

	@Override
	protected void onResume() {
		super.onResume();
		this.getEbitenView().resumeGame();
	}
}