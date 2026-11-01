<script lang="ts">
	import { resolve } from '$app/paths';
</script>

<main>
	<h1>Diary</h1>
	<ul>
		<li><a href={resolve('/foods')}>Foods</a></li>
		<li><a href={resolve('/nutrients')}>Nutrients</a></li>
	</ul>
</main>
